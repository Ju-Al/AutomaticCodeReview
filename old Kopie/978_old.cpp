/**
 * Copyright Soramitsu Co., Ltd. 2017 All Rights Reserved.
 * http://soramitsu.co.jp
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include <backend/protobuf/common_objects/peer.hpp>
#include <boost/accumulators/accumulators.hpp>
#include <boost/math/distributions/chi_squared.hpp>
#include <boost/range/adaptors.hpp>
#include <boost/range/counting_range.hpp>
#include <boost/range/numeric.hpp>
#include <iostream>
#include <unordered_map>
#include "consensus/yac/impl/peer_orderer_impl.hpp"
#include "consensus/yac/storage/yac_proposal_storage.hpp"
#include "model/account.hpp"
#include "model/account_asset.hpp"
#include "model/asset.hpp"
#include "model/domain.hpp"
#include "module/irohad/ametsuchi/ametsuchi_mocks.hpp"
#include "module/irohad/consensus/yac/yac_mocks.hpp"
#include "primitive.pb.h"

using namespace boost::adaptors;
using namespace iroha::ametsuchi;
using namespace iroha::consensus::yac;

using namespace std;
using ::testing::Return;

size_t N_PEERS = 4;
class YacPeerOrdererTest : public ::testing::Test {
 public:
  YacPeerOrdererTest() : orderer(make_shared<MockPeerQuery>()) {}

  void SetUp() override {
    wsv = make_shared<MockPeerQuery>();
    orderer = PeerOrdererImpl(wsv);
  }

  std::vector<iroha::model::Peer> peers = [] {
    std::vector<iroha::model::Peer> result;
    for (size_t i = 1; i <= N_PEERS; ++i) {
      result.push_back(iroha::consensus::yac::mk_peer(std::to_string(i)));
    }
    return result;
  }();

  std::vector<wPeer> s_peers = [] {
    std::vector<wPeer> result;
    for (size_t i = 1; i <= N_PEERS; ++i) {
      auto tmp = iroha::consensus::yac::mk_peer(std::to_string(i));
      iroha::protocol::Peer peer;
      peer.set_address(tmp.address);
      peer.set_peer_key(tmp.pubkey.to_string());
      auto curr =
          shared_model::detail::makePolymorphic<shared_model::proto::Peer>(
              shared_model::proto::Peer(std::move(peer)));
      result.emplace_back(curr);
    }
    return result;
  }();

  shared_ptr<MockPeerQuery> wsv;
  PeerOrdererImpl orderer;
};

TEST_F(YacPeerOrdererTest, PeerOrdererInitialOrderWhenInvokeNormalCase) {
  cout << "----------| InitialOrder() => valid object |----------" << endl;

  EXPECT_CALL(*wsv, getLedgerPeers()).WillOnce(Return(s_peers));
  auto order = orderer.getInitialOrdering();
  ASSERT_EQ(order.value().getPeers(), peers);
}

TEST_F(YacPeerOrdererTest, PeerOrdererInitialOrderWhenInvokeFailCase) {
  cout << "----------| InitialOrder() => nullopt case |----------" << endl;

  EXPECT_CALL(*wsv, getLedgerPeers()).WillOnce(Return(nonstd::nullopt));
  auto order = orderer.getInitialOrdering();
  ASSERT_EQ(order, nonstd::nullopt);
}

TEST_F(YacPeerOrdererTest, PeerOrdererOrderingWhenInvokeNormalCase) {
  cout << "----------| Order() => valid object |----------" << endl;

  EXPECT_CALL(*wsv, getLedgerPeers()).WillOnce(Return(s_peers));
  auto order = orderer.getOrdering(YacHash());
  ASSERT_EQ(order.value().getPeers().size(), peers.size());
}

TEST_F(YacPeerOrdererTest, PeerOrdererOrderingWhenInvokeFaillCase) {
  cout << "----------| Order() => nullopt case |----------" << endl;

  EXPECT_CALL(*wsv, getLedgerPeers()).WillOnce(Return(nonstd::nullopt));
  auto order = orderer.getOrdering(YacHash());
  ASSERT_EQ(order, nonstd::nullopt);
}

/**
 * @given initial peer list in the ledger
 * @when calling ordering function on hash ["1" to "N"] k times
 * @test histogram must be from uniform distibution
 */
TEST_F(YacPeerOrdererTest, FairnessTest) {
  // Calculate number of permutations of peers
  double comb = std::tgamma(N_PEERS + 1);
  // Run experiments N times for each combination
  double exp_val = 30;
  int times = comb * exp_val;
  std::unordered_map<std::string, int> histogram;
  EXPECT_CALL(*wsv, getLedgerPeers())
      .Times(times)
      .WillRepeatedly(Return(s_peers));

  auto peers_set =
      transform(boost::counting_range(1, times + 1), [this](const auto &i) {
        std::string hash = std::to_string(i);
        return orderer.getOrdering(YacHash(hash, hash)).value().getPeers();
      });
  for (const auto &peers : peers_set) {
    std::string res = std::accumulate(peers.begin(),
                                      peers.end(),
                                      std::string(),
                                      [](std::string res, const auto &peer) {
                                        return res + " " + peer.address;
                                      });
    histogram[res]++;
  }
  // chi square goodness of fit test
  auto chi_square = boost::accumulate(
      histogram | map_values, 0.0, [&exp_val](auto stat, const auto &val) {
        return stat + (val - exp_val) * (val - exp_val) / exp_val;
      });
  boost::math::chi_squared chi_dist(comb - 1);
  auto p = boost::math::cdf(chi_dist, chi_square);
  // Significance level 0.05 is used
  ASSERT_LT(0.05, p) << "The distibution is not uniform, p = "
                     << std::to_string(p);
  ASSERT_EQ(comb, histogram.size());
}