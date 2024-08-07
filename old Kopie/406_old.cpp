// Copyright 2016 Proyectos y Sistemas de Mantenimiento SL (eProsima).


ReaderProxy::ReaderProxy(const RemoteReaderAttributes& rdata, const WriterTimes& times, StatefulWriter* SW) :
    m_att(rdata), mp_SFW(SW),
    mp_nackResponse(nullptr), mp_nackSupression(nullptr), m_lastAcknackCount(0),
    mp_mutex(new std::recursive_mutex()), lastNackfragCount_(0)
    if (rdata.endpoint.reliabilityKind == RELIABLE)
    {
        mp_nackResponse = new NackResponseDelay(this, TimeConv::Time_t2MilliSecondsDouble(times.nackResponseDelay));
        mp_nackSupression = new NackSupressionDuration(this, TimeConv::Time_t2MilliSecondsDouble(times.nackSupressionDuration));
    }
    // Use remoteLocatorList as joint unicast + multicast locators
    m_att.endpoint.remoteLocatorList.assign(m_att.endpoint.unicastLocatorList);
    m_att.endpoint.remoteLocatorList.push_back(m_att.endpoint.multicastLocatorList);
    logInfo(RTPS_WRITER, "Reader Proxy created");
ReaderProxy::~ReaderProxy()
    destroy_timers();
    delete(mp_mutex);
void ReaderProxy::destroy_timers()
    if (mp_nackResponse != nullptr)
        delete(mp_nackResponse);
        mp_nackResponse = nullptr;
    if (mp_nackSupression != nullptr)
    {
        delete(mp_nackSupression);
        mp_nackSupression = nullptr;
    }
void ReaderProxy::addChange(const ChangeForReader_t& change)
    std::lock_guard<std::recursive_mutex> guard(*mp_mutex);
    assert(change.getSequenceNumber() > changesFromRLowMark_);
    assert(m_changesForReader.rbegin() != m_changesForReader.rend() ?
        change.getSequenceNumber() > m_changesForReader.rbegin()->getSequenceNumber() :
        true);
    if (m_changesForReader.empty() && change.getStatus() == ACKNOWLEDGED)
        changesFromRLowMark_ = change.getSequenceNumber();
    m_changesForReader.insert(change);
    //TODO (Ricardo) Remove this functionality from here. It is not his place.
    if (change.getStatus() == UNSENT)
        AsyncWriterThread::wakeUp(mp_SFW);
size_t ReaderProxy::countChangesForReader() const
    std::lock_guard<std::recursive_mutex> guard(*mp_mutex);
    return m_changesForReader.size();
bool ReaderProxy::change_is_acked(const SequenceNumber_t& sequence_number)
    std::lock_guard<std::recursive_mutex> guard(*mp_mutex);

    if (sequence_number <= changesFromRLowMark_)
    auto chit = m_changesForReader.find(ChangeForReader_t(sequence_number));
    assert(chit != m_changesForReader.end());
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * @file ReaderProxy.cpp
 *
 */


#include <fastrtps/log/Log.h>
#include <fastrtps/rtps/history/WriterHistory.h>
#include <fastrtps/rtps/resources/AsyncWriterThread.h>
#include <fastrtps/rtps/writer/ReaderProxy.h>
#include <fastrtps/rtps/writer/StatefulWriter.h>
#include <fastrtps/rtps/writer/timedevent/NackResponseDelay.h>
#include <fastrtps/rtps/writer/timedevent/NackSupressionDuration.h>
#include <fastrtps/utils/TimeConversion.h>

#include <mutex>
#include <cassert>
#include <algorithm>

#include "../history/HistoryAttributesExtension.hpp"

namespace eprosima {
namespace fastrtps {
namespace rtps {

ReaderProxy::ReaderProxy(
        const WriterTimes& times, 
        StatefulWriter* writer) 
    : is_active_(false)
    , reader_attributes_()
    , writer_(writer)
    , changes_for_reader_(resource_limits_from_history(writer->mp_history->m_att, 0))
    , nack_response_event_(nullptr)
    , nack_supression_event_(nullptr)
    , timers_enabled_(false)
    , last_acknack_count_(0)
    , last_nackfrag_count_(0)
{
    nack_response_event_ = std::make_shared<NackResponseDelay>(writer_,
        TimeConv::Time_t2MilliSecondsDouble(times.nackResponseDelay));
    nack_supression_event_ = std::make_shared <NackSupressionDuration>(writer_,
        TimeConv::Time_t2MilliSecondsDouble(times.nackSupressionDuration));

    stop();
}

ReaderProxy::~ReaderProxy()
{
}

void ReaderProxy::start(const RemoteReaderAttributes& reader_attributes)
{
    is_active_ = true;
    reader_attributes_ = reader_attributes;

    reader_attributes_.endpoint.remoteLocatorList.assign(reader_attributes_.endpoint.unicastLocatorList);
    reader_attributes_.endpoint.remoteLocatorList.push_back(reader_attributes_.endpoint.multicastLocatorList);

    nack_response_event_->reader_guid(reader_attributes_.guid);
    nack_supression_event_->reader_guid(reader_attributes_.guid);
    timers_enabled_ = (reader_attributes_.endpoint.reliabilityKind == RELIABLE);

    logInfo(RTPS_WRITER, "Reader Proxy started");
}

void ReaderProxy::stop()
{
    is_active_ = false;
    reader_attributes_.guid = c_Guid_Unknown;
    disable_timers();

    changes_for_reader_.clear();
    last_acknack_count_ = 0;
    last_nackfrag_count_ = 0;
    changes_low_mark_ = SequenceNumber_t();
}

void ReaderProxy::disable_timers()
{
    if (timers_enabled_.exchange(false))
    {
        nack_response_event_->cancel_timer();
        nack_supression_event_->cancel_timer();
    }
}

void ReaderProxy::update_nack_response_interval(const Duration_t& interval)
{
    nack_response_event_->update_interval(interval);
}

void ReaderProxy::update_nack_supression_interval(const Duration_t& interval)
{
    nack_supression_event_->update_interval(interval);
}

void ReaderProxy::add_change(
        const ChangeForReader_t& change, 
        bool restart_nack_supression)
{
    assert(change.getSequenceNumber() > changes_low_mark_);
    assert(changes_for_reader_.empty() ? true :
        change.getSequenceNumber() > changes_for_reader_.back().getSequenceNumber());

    if (restart_nack_supression && timers_enabled_)
    {
        nack_supression_event_->restart_timer();
    }

    // For best effort readers, changes are acked when being sent
    if (changes_for_reader_.empty() && change.getStatus() == ACKNOWLEDGED)
    {
        changes_low_mark_ = change.getSequenceNumber();
        return;
    }

    if (changes_for_reader_.push_back(change) == nullptr)
    {
        // This should never happen
        assert(false);
        logError(RTPS_WRITER, "Error adding change " << change.getSequenceNumber() << " to reader proxy " << \
            reader_attributes_.guid);
    }
}

bool ReaderProxy::has_changes() const
{
    return !changes_for_reader_.empty();
}

bool ReaderProxy::change_is_acked(const SequenceNumber_t& seq_num) const
{
    if (seq_num <= changes_low_mark_ || changes_for_reader_.empty())
    {
        return true;
    }

    assert(changes_for_reader_.back().getSequenceNumber() >= seq_num);
    ChangeConstIterator chit = find_change(seq_num);
    if (chit == changes_for_reader_.end())
    {
        // There is a hole in changes_for_reader_
        // This means a change was removed.
        // The case is equivalent to the !chit->isRelevant() code below
        return true;
    }

    return !chit->isRelevant() || chit->getStatus() == ACKNOWLEDGED;
}

void ReaderProxy::acked_changes_set(const SequenceNumber_t& seq_num)
{
    SequenceNumber_t future_low_mark = seq_num;

    if (seq_num > changes_low_mark_)
    {
        ChangeIterator chit = find_change(seq_num);
        changes_for_reader_.erase(changes_for_reader_.begin(), chit);
    }
    else
    {
        // Special case. Currently only used on Builtin StatefulWriters
        // after losing lease duration.

        SequenceNumber_t current_sequence = seq_num;
        SequenceNumber_t min_sequence = writer_->get_seq_num_min();
        if (seq_num < min_sequence)
        {
            current_sequence = min_sequence;
        }
        future_low_mark = current_sequence;

        bool should_sort = false;
        for (; current_sequence <= changes_low_mark_; ++current_sequence)
        {
            // Skip all consecutive changes already in the collection
            ChangeConstIterator it = find_change(current_sequence);
            while( it != changes_for_reader_.end() && 
                current_sequence <= changes_low_mark_ &&
                it->getSequenceNumber() == current_sequence)
            {
                ++current_sequence;
                ++it;
            }

            if (current_sequence <= changes_low_mark_)
            {
                CacheChange_t* change = nullptr;
                if (writer_->mp_history->get_change(current_sequence, writer_->getGuid(), &change))
                {
                    should_sort = true;
                    ChangeForReader_t cr(change);
                    cr.setStatus(UNACKNOWLEDGED);
                    changes_for_reader_.push_back(cr);
                }
            }
        }

        // Keep changes sorted by sequence number
        if (should_sort)
        {
            std::sort(changes_for_reader_.begin(), changes_for_reader_.end(), ChangeForReaderCmp());
        }
    }

    changes_low_mark_ = future_low_mark - 1;
}

bool ReaderProxy::requested_changes_set(const SequenceNumberSet_t& seq_num_set)
{
    bool isSomeoneWasSetRequested = false;

    seq_num_set.for_each([&](SequenceNumber_t sit)
    {
        ChangeIterator chit = find_change(sit);
        if (chit != changes_for_reader_.end())
        {
            if (chit->getStatus() == UNACKNOWLEDGED)
            {
                chit->setStatus(REQUESTED);
                chit->markAllFragmentsAsUnsent();
                isSomeoneWasSetRequested = true;
            }
        }
    });

    if (isSomeoneWasSetRequested)
    {
        logInfo(RTPS_WRITER, "Requested Changes: " << seq_num_set);
        if (timers_enabled_)
        {
            nack_response_event_->restart_timer();
        }
    }
    else if (!seq_num_set.empty())
    {
        logWarning(RTPS_WRITER, "Requested Changes: " << seq_num_set
            << " not found (low mark: " << changes_low_mark_ << ")");
    }

    return isSomeoneWasSetRequested;
}

bool ReaderProxy::set_change_to_status(
        const SequenceNumber_t& seq_num, 
        ChangeForReaderStatus_t status,
        bool restart_nack_supression)
{
    if (restart_nack_supression && is_reliable())
    {
        assert(timers_enabled_);
        nack_supression_event_->restart_timer();
    }

    if (seq_num <= changes_low_mark_)
    {
        return false;
    }

    auto it = find_change(seq_num);
    bool change_was_modified = false;

    // If the change following the low mark is acknowledged, low mark is advanced.
    // Note that this could be the first change in the collection or a hole if the
    // first unacknowledged change is irrelevant.
    if (status == ACKNOWLEDGED && seq_num == changes_low_mark_ + 1)
    {
        changes_low_mark_ = seq_num;
        change_was_modified = true;
    }

    if (it != changes_for_reader_.end())
    {
        if (status == ACKNOWLEDGED && changes_low_mark_ == seq_num)
        {
            // Erase the first change when it is acknowledged
            assert(it == changes_for_reader_.begin());
            changes_for_reader_.erase(it);
        }
        else
        {
            // Otherwise change status
            if (it->getStatus() != status)
            {
                it->setStatus(status);
                change_was_modified = true;
            }
        }
    }
    
    return change_was_modified;
}

bool ReaderProxy::mark_fragment_as_sent_for_change(
        const SequenceNumber_t& seq_num, 
        FragmentNumber_t frag_num,
        bool& was_last_fragment)
{
    was_last_fragment = false;

    if (seq_num <= changes_low_mark_)
    {
        return false;
    }

    bool change_found = false;
    auto it = find_change(seq_num);

    if (it != changes_for_reader_.end())
    {
        change_found = true;
        it->markFragmentsAsSent(frag_num);
        was_last_fragment = it->getUnsentFragments().empty();
    }

    return change_found;

}

bool ReaderProxy::perform_nack_supression()
{
    return convert_status_on_all_changes(UNDERWAY, UNACKNOWLEDGED);
}

bool ReaderProxy::perform_acknack_response()
{
    return convert_status_on_all_changes(REQUESTED, UNSENT);
}

bool ReaderProxy::convert_status_on_all_changes(
        ChangeForReaderStatus_t previous, 
        ChangeForReaderStatus_t next)
{
    assert(previous != next);

    // NOTE: This is only called for REQUESTED=>UNSENT (acknack response) or 
    //       UNDERWAY=>UNACKNOWLEDGED (nack supression)

    bool at_least_one_modified = false;
    for(ChangeForReader_t& change : changes_for_reader_)
    {
        if (change.getStatus() == previous)
        {
            at_least_one_modified = true;
            change.setStatus(next);
        }
    }

    return at_least_one_modified;
}

void ReaderProxy::change_has_been_removed(const SequenceNumber_t& seq_num)
{
    // Check sequence number is in the container, because it was not clean up.
    if (changes_for_reader_.empty() || seq_num < changes_for_reader_.begin()->getSequenceNumber())
    {
        return;
    }

    // Element may not be in the container when marked as irrelevant.
    auto chit = find_change(seq_num);
    changes_for_reader_.erase(chit);
}

bool ReaderProxy::has_unacknowledged() const
{
    for (const ChangeForReader_t& it : changes_for_reader_)
    {
        if (it.isRelevant() && it.getStatus() == UNACKNOWLEDGED)
        {
            return true;
        }
    }

    return false;
}

bool ReaderProxy::requested_fragment_set(
        const SequenceNumber_t& seq_num, 
        const FragmentNumberSet_t& frag_set)
{
    // Locate the outbound change referenced by the NACK_FRAG
    ChangeIterator changeIter = find_change(seq_num);
    if (changeIter == changes_for_reader_.end())
    {
        return false;
    }

    changeIter->markFragmentsAsUnsent(frag_set);

    // If it was UNSENT, we shouldn't switch back to REQUESTED to prevent stalling.
    if (changeIter->getStatus() != UNSENT)
    {
        changeIter->setStatus(REQUESTED);
    }

    return true;
}

bool ReaderProxy::process_nack_frag(
        const GUID_t& reader_guid, 
        uint32_t nack_count,
        const SequenceNumber_t& seq_num,
        const FragmentNumberSet_t& fragments_state)
{
    if (reader_attributes_.guid == reader_guid)
    {
        if (last_nackfrag_count_ < nack_count)
        {
            last_nackfrag_count_ = nack_count;
            // TODO Not doing Acknowledged.
            if (requested_fragment_set(seq_num, fragments_state))
            {
                nack_response_event_->restart_timer();
            }
        }

        return true;
    }

    return false;
}

static bool change_less_than_sequence(
    const ChangeForReader_t& change,
    const SequenceNumber_t& seq_num)
{
    return change.getSequenceNumber() < seq_num;
}

ReaderProxy::ChangeIterator ReaderProxy::find_change(const SequenceNumber_t& seq_num)
{
    ReaderProxy::ChangeIterator it;
    ReaderProxy::ChangeIterator end = changes_for_reader_.end();
    it = std::lower_bound(changes_for_reader_.begin(), end, seq_num, change_less_than_sequence);
    
    return it == end 
        ? it 
        : it->getSequenceNumber() == seq_num ? it : end;
}

ReaderProxy::ChangeConstIterator ReaderProxy::find_change(const SequenceNumber_t& seq_num) const
{
    ReaderProxy::ChangeConstIterator it;
    ReaderProxy::ChangeConstIterator end = changes_for_reader_.end();
    it = std::lower_bound(changes_for_reader_.begin(), end, seq_num, change_less_than_sequence);

    return it == end
        ? it
        : it->getSequenceNumber() == seq_num ? it : end;
}

}   // namespace rtps
}   // namespace fastrtps
}   // namespace eprosima
