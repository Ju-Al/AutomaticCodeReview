/**
 * Copyright Soramitsu Co., Ltd. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

#include "consensus/yac/storage/yac_block_storage.hpp"

#include <gtest/gtest.h>
#include <boost/optional.hpp>
#include "consensus/yac/storage/yac_proposal_storage.hpp"
#include "consensus/yac/storage/yac_vote_storage.hpp"
#include "logger/logger.hpp"
#include "module/irohad/consensus/yac/yac_mocks.hpp"

using namespace iroha::consensus::yac;

static logger::Logger log_ = logger::testLog("YacBlockStorage");

class YacBlockStorageTest : public ::testing::Test {
 public:
  YacHash hash;
  PeersNumberType number_of_peers;
  YacBlockStorage storage = YacBlockStorage(
      YacHash(iroha::consensus::Round{1, 1}, "proposal", "commit"), 4);
  PeersNumberType supermajority = number_of_peers - number_of_peers / 5;
  YacHash hash = YacHash(iroha::consensus::Round{1, 1}, "proposal", "commit");
  YacBlockStorage storage = YacBlockStorage(hash, number_of_peers);
  std::vector<VoteMessage> valid_votes;

  void SetUp() override {
    valid_votes.reserve(number_of_peers);
    std::generate_n(std::back_inserter(valid_votes), number_of_peers, [this] {
      return create_vote(this->hash, std::to_string(this->valid_votes.size()));
    });
  }
};

TEST_F(YacBlockStorageTest, YacBlockStorageWhenNormalDataInput) {
  log_->info("------------| Sequential insertion of votes |------------");

  for (size_t num_inserted = 0; num_inserted < number_of_peers;) {
    auto insert_result = storage.insert(valid_votes.at(num_inserted++));
    if (num_inserted < supermajority) {
      EXPECT_EQ(boost::none, insert_result)
          << "Got an Answer after inserting " << num_inserted
          << " votes, before the supermajority reached at " << supermajority;
    } else {
      // after supermajority reached we have a CommitMessage
      ASSERT_NE(boost::none, insert_result)
          << "Did not get an Answer after inserting " << num_inserted
          << " votes and the supermajority reached at " << supermajority;
      ASSERT_EQ(num_inserted,
                boost::get<CommitMessage>(*insert_result).votes.size())
          << " The commit message must have all the previously inserted votes.";
    }
  }
}

TEST_F(YacBlockStorageTest, YacBlockStorageWhenNotCommittedAndCommitAcheive) {
  log_->info("-----------| Insert vote => insert commit |-----------");

  auto insert_1 = storage.insert(valid_votes.at(0));
  ASSERT_EQ(boost::none, insert_1);

  decltype(YacBlockStorageTest::valid_votes) for_insert(valid_votes.begin() + 1,
                                                        valid_votes.end());
  auto insert_commit = storage.insert(for_insert);
  ASSERT_EQ(number_of_peers,
            boost::get<CommitMessage>(*insert_commit).votes.size());
}

TEST_F(YacBlockStorageTest, YacBlockStorageWhenGetVotes) {
  log_->info("-----------| Init storage => verify internal votes |-----------");

  storage.insert(valid_votes);
  ASSERT_EQ(valid_votes, storage.getVotes());
}

TEST_F(YacBlockStorageTest, YacBlockStorageWhenIsContains) {
  log_->info(
      "-----------| Init storage => "
      "verify ok and fail cases of contains |-----------");

  decltype(YacBlockStorageTest::valid_votes) for_insert(
      valid_votes.begin(), valid_votes.begin() + 2);

  storage.insert(for_insert);

  ASSERT_TRUE(storage.isContains(valid_votes.at(0)));
  ASSERT_FALSE(storage.isContains(valid_votes.at(3)));
}