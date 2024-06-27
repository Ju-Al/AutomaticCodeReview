/******************************************************************************
          && builder_->add(std::string_view{str, packet_size}))) {
 *                    _   _____   __________                                  *
 *                   | | / / _ | / __/_  __/     Visibility                   *
 *                   | |/ / __ |_\ \  / /          Across                     *
 *                   |___/_/ |_/___/ /_/       Space and Time                 *
 *                                                                            *
 * This file is part of VAST. It is subject to the license terms in the       *
 * LICENSE file found in the top-level directory of this distribution and at  *
 * http://vast.io/license. No part of VAST, including this file, may be       *
 * copied, modified, propagated, or distributed except according to the terms *
 * contained in the LICENSE file.                                             *
 ******************************************************************************/

#include <netinet/in.h>

#include <thread>

#include "vast/community_id.hpp"
#include "vast/detail/assert.hpp"
#include "vast/detail/byte_swap.hpp"
#include "vast/error.hpp"
#include "vast/event.hpp"
#include "vast/filesystem.hpp"
#include "vast/format/pcap.hpp"
#include "vast/logger.hpp"
#include "vast/table_slice.hpp"
#include "vast/table_slice_builder.hpp"

namespace vast {
namespace format {
namespace pcap {
namespace {

inline type make_packet_type() {
  return record_type{{"time", time_type{}.attributes({{"timestamp"}})},
                     {"src", address_type{}},
                     {"dst", address_type{}},
                     {"sport", port_type{}},
                     {"dport", port_type{}},
                     {"payload", string_type{}.attributes({{"skip"}})},
                     {"community_id", string_type{}}}
    .name("pcap.packet");
}

auto const pcap_packet_type = make_packet_type();

} // namespace <anonymous>

reader::reader(caf::atom_value id, const caf::settings& /*options*/,
               std::string input, uint64_t cutoff, size_t max_flows,
               size_t max_age, size_t expire_interval, int64_t pseudo_realtime)
  : super(id),
    packet_type_{pcap_packet_type},
    cutoff_{cutoff},
    max_flows_{max_flows},
    max_age_{max_age},
    expire_interval_{expire_interval},
    pseudo_realtime_{pseudo_realtime},
    input_{std::move(input)} {
  // nop
}

reader::~reader() {
  if (pcap_)
    ::pcap_close(pcap_);
}

caf::error reader::schema(vast::schema sch) {
  return replace_if_congruent({&packet_type_}, sch);
}

schema reader::schema() const {
  vast::schema result;
  result.add(packet_type_);
  return result;
}

const char* reader::name() const {
  return "pcap-reader";
}

caf::error reader::read_impl(size_t max_events, size_t max_slice_size,
                             consumer& f) {
  // Sanity checks.
  VAST_ASSERT(max_events > 0);
  VAST_ASSERT(max_slice_size > 0);
  if (builder_ == nullptr) {
    if (!caf::holds_alternative<record_type>(packet_type_))
      return make_error(ec::parse_error, "illegal packet type");
    if (!reset_builder(caf::get<record_type>(packet_type_)))
      return make_error(ec::parse_error,
                        "unable to create builder for packet type");
  }
  // Local buffer for storing error messages.
  char buf[PCAP_ERRBUF_SIZE];
  // Initialize PCAP if needed.
  if (!pcap_) {
    // Determine interfaces.
    pcap_if_t* iface;
    if (::pcap_findalldevs(&iface, buf) == -1)
      return make_error(ec::format_error,
                        "failed to enumerate interfaces: ", buf);
    for (auto i = iface; i != nullptr; i = i->next)
      if (input_ == i->name) {
        pcap_ = ::pcap_open_live(i->name, 65535, 1, 1000, buf);
        if (!pcap_) {
          ::pcap_freealldevs(iface);
          return make_error(ec::format_error, "failed to open interface ",
                            input_, ": ", buf);
        }
        if (pseudo_realtime_ > 0) {
          pseudo_realtime_ = 0;
          VAST_WARNING(this, "ignores pseudo-realtime in live mode");
        }
        VAST_DEBUG(this, "listens on interface " << i->name);
        break;
      }
    ::pcap_freealldevs(iface);
    if (!pcap_) {
      if (input_ != "-" && !exists(input_))
        return make_error(ec::format_error, "no such file: ", input_);
#ifdef PCAP_TSTAMP_PRECISION_NANO
      pcap_ = ::
        pcap_open_offline_with_tstamp_precision(input_.c_str(),
                                                PCAP_TSTAMP_PRECISION_NANO,
                                                buf);
#else
      pcap_ = ::pcap_open_offline(input_.c_str(), buf);
#endif
      if (!pcap_) {
        flows_.clear();
        return make_error(ec::format_error, "failed to open pcap file ", input_,
                          ": ", std::string{buf});
      }
      VAST_DEBUG(this, "reads trace from", input_);
      if (pseudo_realtime_ > 0)
        VAST_DEBUG(this, "uses pseudo-realtime factor 1/" << pseudo_realtime_);
    }
    VAST_DEBUG(this, "cuts off flows after", cutoff_,
               "bytes in each direction");
    VAST_DEBUG(this, "keeps at most", max_flows_, "concurrent flows");
    VAST_DEBUG(this, "evicts flows after", max_age_ << "s of inactivity");
    VAST_DEBUG(this, "expires flow table every", expire_interval_ << "s");
  }
  for (size_t produced = 0; produced < max_events; ++produced) {
    const uint8_t* data;
    pcap_pkthdr* header;
    auto r = ::pcap_next_ex(pcap_, &header, &data);
    if (r == 0) {
      // Attempt to fetch next packet timed out.
      return finish(f, caf::none);
    }
    if (r == -2)
      return finish(f, make_error(ec::end_of_input, "reached end of trace"));
    if (r == -1) {
      auto err = std::string{::pcap_geterr(pcap_)};
      ::pcap_close(pcap_);
      pcap_ = nullptr;
      return finish(f, make_error(ec::format_error,
                                  "failed to get next packet: ", err));
    }
    // Parse packet.
    connection conn;
    auto packet_size = header->len - 14;
    auto layer3 = data + 14;
    const uint8_t* layer4 = nullptr;
    uint8_t layer4_proto = 0;
    auto layer2_type = *reinterpret_cast<const uint16_t*>(data + 12);
    uint64_t payload_size = packet_size;
    switch (detail::to_host_order(layer2_type)) {
      default:
        // Skip all non-IP packets.
        continue;
      case 0x0800: {
        if (header->len < 14 + 20)
          return make_error(ec::format_error, "IPv4 header too short");
        size_t header_size = (*layer3 & 0x0f) * 4;
        if (header_size < 20)
          return make_error(ec::format_error,
                            "IPv4 header too short: ", header_size, " bytes");
        auto orig_h = reinterpret_cast<const uint32_t*>(layer3 + 12);
        auto resp_h = reinterpret_cast<const uint32_t*>(layer3 + 16);
        conn.src = {orig_h, address::ipv4, address::network};
        conn.dst = {resp_h, address::ipv4, address::network};
        layer4_proto = *(layer3 + 9);
        layer4 = layer3 + header_size;
        payload_size -= header_size;
        break;
      }
      case 0x86dd: {
        if (header->len < 14 + 40)
          return make_error(ec::format_error, "IPv6 header too short");
        auto orig_h = reinterpret_cast<const uint32_t*>(layer3 + 8);
        auto resp_h = reinterpret_cast<const uint32_t*>(layer3 + 24);
        conn.src = {orig_h, address::ipv4, address::network};
        conn.dst = {resp_h, address::ipv4, address::network};
        layer4_proto = *(layer3 + 6);
        layer4 = layer3 + 40;
        payload_size -= 40;
        break;
      }
    }
    if (layer4_proto == IPPROTO_TCP) {
      VAST_ASSERT(layer4);
      auto orig_p = *reinterpret_cast<const uint16_t*>(layer4);
      auto resp_p = *reinterpret_cast<const uint16_t*>(layer4 + 2);
      orig_p = detail::to_host_order(orig_p);
      resp_p = detail::to_host_order(resp_p);
      conn.sport = {orig_p, port::tcp};
      conn.dport = {resp_p, port::tcp};
      auto data_offset = *reinterpret_cast<const uint8_t*>(layer4 + 12) >> 4;
      payload_size -= data_offset * 4;
    } else if (layer4_proto == IPPROTO_UDP) {
      VAST_ASSERT(layer4);
      auto orig_p = *reinterpret_cast<const uint16_t*>(layer4);
      auto resp_p = *reinterpret_cast<const uint16_t*>(layer4 + 2);
      orig_p = detail::to_host_order(orig_p);
      resp_p = detail::to_host_order(resp_p);
      conn.sport = {orig_p, port::udp};
      conn.dport = {resp_p, port::udp};
      payload_size -= 8;
    } else if (layer4_proto == IPPROTO_ICMP) {
      VAST_ASSERT(layer4);
      auto message_type = *reinterpret_cast<const uint8_t*>(layer4);
      auto message_code = *reinterpret_cast<const uint8_t*>(layer4 + 1);
      conn.sport = {message_type, port::icmp};
      conn.dport = {message_code, port::icmp};
      payload_size -= 8; // TODO: account for variable-size data.
    }
    // Parse packet timestamp
    uint64_t packet_time = header->ts.tv_sec;
    if (last_expire_ == 0)
      last_expire_ = packet_time;
    if (!update_flow(conn, packet_time, payload_size)) {
      // Skip cut off packets.
      continue;
    }
    evict_inactive(packet_time);
    shrink_to_max_size();
    // Extract timestamp.
    using namespace std::chrono;
    auto secs = seconds(header->ts.tv_sec);
    auto ts = time{duration_cast<duration>(secs)};
#ifdef PCAP_TSTAMP_PRECISION_NANO
    ts += nanoseconds(header->ts.tv_usec);
#else
    ts += microseconds(header->ts.tv_usec);
#endif
    // Assemble packet.
    // We start with the network layer and skip the link layer.
    auto str = reinterpret_cast<const char*>(data + 14);
    auto& cid = flow(conn).community_id;
    if (!(builder_->add(ts) && builder_->add(conn.src)
          && builder_->add(conn.dst) && builder_->add(conn.sport)
          && builder_->add(conn.dport)
          && builder_->add(std::string_view{str, packet_size})
          && builder_->add(std::string_view(cid)))) {
      return make_error(ec::parse_error, "unable to fill row");
    }
    if (pseudo_realtime_ > 0) {
      if (ts < last_timestamp_) {
        VAST_WARNING(this, "encountered non-monotonic packet timestamps:",
                     ts.time_since_epoch().count(), '<',
                     last_timestamp_.time_since_epoch().count());
      }
      if (last_timestamp_ != time::min()) {
        auto delta = ts - last_timestamp_;
        std::this_thread::sleep_for(delta / pseudo_realtime_);
      }
      last_timestamp_ = ts;
    }
    if (builder_->rows() == max_slice_size)
      if (auto err = finish(f, caf::none))
        return err;
  }
  return finish(f, caf::none);
}

reader::connection_state& reader::flow(const connection& conn) {
  auto i = flows_.find(conn);
  if (i == flows_.end()) {
    auto cf = community_id::flow{conn.src, conn.dst, conn.sport, conn.dport};
    auto id = community_id::compute<policy::base64>(cf);
    i = flows_.emplace(conn, connection_state{0, 0, std::move(id)}).first;
  }
  return i->second;
}

bool reader::update_flow(const connection& conn, uint64_t packet_time,
                         uint64_t payload_size) {
  auto& packet_flow = flow(conn);
  packet_flow.last = packet_time;
  auto& flow_size = packet_flow.bytes;
  if (flow_size == cutoff_)
    return false;
  VAST_ASSERT(flow_size < cutoff_);
  // Trim the packet if needed.
  flow_size += std::min(payload_size, cutoff_ - flow_size);
  return true;
}

void reader::evict_inactive(uint64_t packet_time) {
  if (packet_time - last_expire_ <= expire_interval_)
    return;
  last_expire_ = packet_time;
  auto i = flows_.begin();
  while (i != flows_.end())
    if (packet_time - i->second.last > max_age_)
      i = flows_.erase(i);
    else
      ++i;
}

void reader::shrink_to_max_size() {
  while (flows_.size() >= max_flows_) {
    auto buckets = flows_.bucket_count();
    auto unif1 = std::uniform_int_distribution<size_t>{0, buckets - 1};
    auto bucket = unif1(generator_);
    auto bucket_size = flows_.bucket_size(bucket);
    while (bucket_size == 0) {
      ++bucket;
      bucket %= buckets;
      bucket_size = flows_.bucket_size(bucket);
    }
    auto unif2 = std::uniform_int_distribution<size_t>{0, bucket_size - 1};
    auto offset = unif2(generator_);
    VAST_ASSERT(offset < bucket_size);
    auto begin = flows_.begin(bucket);
    std::advance(begin, offset);
    flows_.erase(begin->first);
  }
}

writer::writer(std::string trace, size_t flush_interval)
  : flush_interval_{flush_interval},
    trace_{std::move(trace)} {
}

writer::~writer() {
  if (dumper_)
    ::pcap_dump_close(dumper_);
  if (pcap_)
    ::pcap_close(pcap_);
}

caf::error writer::write(const table_slice& slice) {
  if (!pcap_) {
#ifdef PCAP_TSTAMP_PRECISION_NANO
    pcap_ = ::pcap_open_dead_with_tstamp_precision(DLT_RAW, 65535,
                                                   PCAP_TSTAMP_PRECISION_NANO);
#else
    pcap_ = ::pcap_open_dead(DLT_RAW, 65535);
#endif
    if (!pcap_)
      return make_error(ec::format_error, "failed to open pcap handle");
    dumper_ = ::pcap_dump_open(pcap_, trace_.c_str());
    if (!dumper_)
      return make_error(ec::format_error, "failed to open pcap dumper");
  }
  if (!congruent(slice.layout(), pcap_packet_type))
    return make_error(ec::format_error, "invalid pcap packet type");
  // TODO: consider iterating in natural order for the slice.
  for (size_t row = 0; row < slice.rows(); ++row) {
    auto payload_field = slice.at(row, 5);
    auto& payload = caf::get<view<std::string>>(payload_field);
    // Make PCAP header.
    ::pcap_pkthdr header;
    auto ns_field = slice.at(row, 0);
    auto ns = caf::get<view<time>>(ns_field).time_since_epoch().count();
    header.ts.tv_sec = ns / 1000000000;
#ifdef PCAP_TSTAMP_PRECISION_NANO
    header.ts.tv_usec = ns % 1000000000;
#else
    ns /= 1000;
    header.ts.tv_usec = ns % 1000000;
#endif
    header.caplen = payload.size();
    header.len = payload.size();
    // Dump packet.
    ::pcap_dump(reinterpret_cast<uint8_t*>(dumper_), &header,
                reinterpret_cast<const uint8_t*>(payload.data()));
  }
  if (++total_packets_ % flush_interval_ == 0)
    if (auto r = flush(); !r)
      return r.error();
  return caf::none;
}

expected<void> writer::flush() {
  if (!dumper_)
    return make_error(ec::format_error, "pcap dumper not open");
  VAST_DEBUG(this, "flushes at packet", total_packets_);
  if (::pcap_dump_flush(dumper_) == -1)
    return make_error(ec::format_error, "failed to flush");
  return no_error;
}

const char* writer::name() const {
  return "pcap-writer";
}

} // namespace pcap
} // namespace format
} // namespace vast
