/**
  const size_t commit_delay = 1000;

  auto ordering_service = initOs(max_proposal, commit_delay);
  fake_transport->subscribe(ordering_service);

  // Init => proposal size 5 => 2 proposals after 10 transactions
  size_t call_count = 0;
      .Times(2)
      .WillRepeatedly(InvokeWithoutArgs([&] {
        ++call_count;
        cv.notify_one();
      }));

  shared_model::proto::PeerBuilder builder;

  auto key = shared_model::crypto::PublicKey(peer->pubkey().toString());
  auto tmp = builder.address(peer->address()).pubkey(key).build();

  for (size_t i = 0; i < 10; ++i) {

  std::unique_lock<std::mutex> lock(m);
  cv.wait_for(lock, 10s, [&] { return call_count == 2; });
  // Init => proposal timer 400 ms => 10 tx by 50 ms => 2 proposals in 1 second
  shared_model::proto::PeerBuilder builder;

  auto key = shared_model::crypto::PublicKey(peer->pubkey().toString());
  auto tmp = builder.address(peer->address()).pubkey(key).build();

  EXPECT_CALL(*wsv, getLedgerPeers())
      .WillRepeatedly(Return(std::vector<decltype(peer)>{peer}));

  const size_t max_proposal = 100;
  const size_t commit_delay = 400;


  auto ordering_service = initOs(max_proposal, commit_delay);
  fake_transport->subscribe(ordering_service);

      .WillRepeatedly(InvokeWithoutArgs([&] {
        log_->info("Proposal send to grpc");
        cv.notify_one();
      }));
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

#include <grpc++/grpc++.h>

#include "backend/protobuf/common_objects/peer.hpp"
#include "builders/protobuf/common_objects/proto_peer_builder.hpp"
#include "builders/protobuf/transaction.hpp"
#include "logger/logger.hpp"
#include "module/irohad/ametsuchi/ametsuchi_mocks.hpp"
#include "module/irohad/network/network_mocks.hpp"
#include "module/irohad/ordering/mock_ordering_service_persistent_state.hpp"
#include "module/shared_model/builders/protobuf/test_proposal_builder.hpp"
#include "module/shared_model/builders/protobuf/test_transaction_builder.hpp"
#include "ordering/impl/ordering_service_impl.hpp"
#include "ordering/impl/ordering_service_transport_grpc.hpp"

using namespace iroha;
using namespace iroha::ordering;
using namespace iroha::network;
using namespace iroha::ametsuchi;
using namespace std::chrono_literals;

using ::testing::_;
using ::testing::AtLeast;
using ::testing::DoAll;
using ::testing::Invoke;
using ::testing::InvokeWithoutArgs;
using ::testing::Return;

static logger::Logger log_ = logger::testLog("OrderingService");

class MockOrderingServiceTransport : public network::OrderingServiceTransport {
 public:
  void subscribe(std::shared_ptr<network::OrderingServiceNotification>
                     subscriber) override {
    subscriber_ = subscriber;
  }

  void publishProposal(
      std::unique_ptr<shared_model::interface::Proposal> proposal,
      const std::vector<std::string> &peers) override {
    return publishProposalProxy(proposal.get(), peers);
  }

  MOCK_METHOD2(publishProposalProxy,
               void(shared_model::interface::Proposal *proposal,
                    const std::vector<std::string> &peers));

  std::weak_ptr<network::OrderingServiceNotification> subscriber_;
};

class OrderingServiceTest : public ::testing::Test {
 public:
  OrderingServiceTest() {
    peer = clone(shared_model::proto::PeerBuilder()
                     .address(address)
                     .pubkey(shared_model::interface::types::PubkeyType(
                         std::string(32, '0')))
                     .build());
  }

  void SetUp() override {
    wsv = std::make_shared<MockPeerQuery>();
    fake_transport = std::make_shared<MockOrderingServiceTransport>();
    fake_persistent_state =
        std::make_shared<MockOrderingServicePersistentState>();
  }

  auto getTx() {
    return std::make_unique<shared_model::proto::Transaction>(
        shared_model::proto::TransactionBuilder()
            .createdTime(iroha::time::now())
            .creatorAccountId("admin@ru")
            .addAssetQuantity("admin@tu", "coin#coin", "1.0")
            .build()
            .signAndAddSignature(
                shared_model::crypto::DefaultCryptoAlgorithmType::
                    generateKeypair()));
  }

  auto initOs(size_t max_proposal) {
    return std::make_shared<OrderingServiceImpl>(
        wsv,
        max_proposal,
        proposal_timeout.get_observable(),
        fake_transport,
        fake_persistent_state,
        false);
  }

  void makeProposalTimeout() {
    log_->info("proposal timeout tick");
    proposal_timeout.get_subscriber().on_next(0);
  }

  std::shared_ptr<MockOrderingServiceTransport> fake_transport;
  std::shared_ptr<MockOrderingServicePersistentState> fake_persistent_state;
  std::condition_variable cv;
  std::mutex m;
  std::string address{"0.0.0.0:50051"};
  std::shared_ptr<shared_model::interface::Peer> peer;
  std::shared_ptr<MockPeerQuery> wsv;
  rxcpp::subjects::subject<long> proposal_timeout;
};

/**
 * @given OrderingService and MockOrderingServiceTransport
 * @when publishProposal is called at transport
 * @then publishProposalProxy is called
 */
TEST_F(OrderingServiceTest, SimpleTest) {
  const size_t max_proposal = 5;

  EXPECT_CALL(*fake_persistent_state, loadProposalHeight())
      .Times(1)
      .WillOnce(Return(boost::optional<size_t>(2)));
  EXPECT_CALL(*fake_transport, publishProposalProxy(_, _)).Times(1);

  auto ordering_service = initOs(max_proposal);
  fake_transport->subscribe(ordering_service);

  fake_transport->publishProposal(
      std::make_unique<shared_model::proto::Proposal>(
          TestProposalBuilder()
              .height(1)
              .createdTime(iroha::time::now())
              .build()),
      {});
}

/**
 * @given OrderingService with max_proposal==5 and only self peer
 *        and MockOrderingServiceTransport
 *        and MockOrderingServicePersistentState
 * @when OrderingService::onTransaction called 10 times
 * @then publishProposalProxy called twice
 *       and proposal height was loaded once and saved twice
 */
TEST_F(OrderingServiceTest, ValidWhenProposalSizeStrategy) {
  const size_t max_proposal = 5;
  const size_t tx_num = 10;

  EXPECT_CALL(*fake_persistent_state, saveProposalHeight(_))
      .Times(2)
      .WillRepeatedly(Return(true));
  EXPECT_CALL(*fake_persistent_state, loadProposalHeight())
      .Times(1)
      .WillOnce(Return(boost::optional<size_t>(2)));
  EXPECT_CALL(*fake_transport, publishProposalProxy(_, _))
      .Times(tx_num / max_proposal);
  EXPECT_CALL(*wsv, getLedgerPeers())
      .WillRepeatedly(Return(std::vector<decltype(peer)>{peer}));

  auto ordering_service = initOs(max_proposal);
  fake_transport->subscribe(ordering_service);

  for (size_t i = 0; i < tx_num; ++i) {
    ordering_service->onTransaction(getTx());
  }
}

/**
 * @given OrderingService with big enough max_proposal and only self peer
 *        and MockOrderingServiceTransport
 *        and MockOrderingServicePersistentState
 * @when OrderingService::onTransaction called 8 times
 *       and after triggered timeout
 *       and then repeat with 2 onTransaction calls
 * @then publishProposalProxy called twice
 *       and proposal height was loaded once and saved twice
 */
TEST_F(OrderingServiceTest, ValidWhenTimerStrategy) {
  const size_t max_proposal = 100;

  EXPECT_CALL(*fake_persistent_state, saveProposalHeight(_))
      .Times(2)
      .WillRepeatedly(Return(true));
  EXPECT_CALL(*fake_persistent_state, loadProposalHeight())
      .Times(1)
      .WillOnce(Return(boost::optional<size_t>(2)));
  EXPECT_CALL(*wsv, getLedgerPeers())
      .WillRepeatedly(Return(std::vector<decltype(peer)>{peer}));
  EXPECT_CALL(*fake_transport, publishProposalProxy(_, _))
      .Times(2)
      .WillRepeatedly(
          InvokeWithoutArgs([&] { log_->info("Proposal send to grpc"); }));

  auto ordering_service = initOs(max_proposal);
  fake_transport->subscribe(ordering_service);

  for (size_t i = 0; i < 8; ++i) {
    ordering_service->onTransaction(getTx());
  }
  makeProposalTimeout();

  ordering_service->onTransaction(getTx());
  ordering_service->onTransaction(getTx());
  makeProposalTimeout();
}

/**
 * @given Ordering service and the persistent state that cannot save
 *        proposals
 * @when onTransaction is called
 * @then no published proposal
 */
TEST_F(OrderingServiceTest, BrokenPersistentState) {
  const size_t max_proposal = 1;

  EXPECT_CALL(*fake_persistent_state, loadProposalHeight())
      .Times(1)
      .WillOnce(Return(boost::optional<size_t>(1)));
  EXPECT_CALL(*fake_persistent_state, saveProposalHeight(2))
      .Times(1)
      .WillRepeatedly(Return(false));

  auto ordering_service = initOs(max_proposal);
  ordering_service->onTransaction(getTx());
  makeProposalTimeout();
}

/**
 * @given Ordering service up and running
 * @when Send 1000 transactions from each of 2 threads
 * @then Ordering service should not crash
 */
TEST_F(OrderingServiceTest, ConcurrentGenerateProposal) {
  const auto max_proposal = 1;
  EXPECT_CALL(*fake_persistent_state, loadProposalHeight())
      .Times(1)
      .WillOnce(Return(boost::optional<size_t>(1)));
  EXPECT_CALL(*fake_persistent_state, saveProposalHeight(_))
      .WillRepeatedly(Return(false));

  auto ordering_service = initOs(max_proposal);

  auto on_tx = [&]() {
    for (int i = 0; i < 1000; ++i) {
      ordering_service->onTransaction(getTx());
    }
  };

  const auto num_threads = 2;

  std::vector<std::thread> threads;
  for (int i = 0; i < num_threads; ++i) {
    threads.emplace_back(std::thread(on_tx));
  }

  for (int i = 0; i < num_threads; ++i) {
    threads.at(i).join();
  }
  makeProposalTimeout();
}

/**
 * @given Ordering service up and running
 * @when Send 1000 transactions from a separate thread and perform 5 second
 * delay during generateProposal() so destructor of OrderingServiceImpl is
 * called before generateProposal() finished
 * @then Ordering service should not crash and publishProposal() should not be
 * called after destructor call
 */
TEST_F(OrderingServiceTest, GenerateProposalDestructor) {
  const auto max_proposal = 100000;
  const auto commit_delay = 100ms;
  EXPECT_CALL(*fake_persistent_state, loadProposalHeight())
      .Times(1)
      .WillOnce(Return(boost::optional<size_t>(1)));
  EXPECT_CALL(*fake_persistent_state, saveProposalHeight(_))
      .WillRepeatedly(InvokeWithoutArgs([] {
        std::this_thread::sleep_for(5s);
        return true;
      }));
  EXPECT_CALL(*wsv, getLedgerPeers())
      .WillRepeatedly(Return(std::vector<decltype(peer)>{peer}));

  {
    EXPECT_CALL(*fake_transport, publishProposalProxy(_, _)).Times(AtLeast(1));
    OrderingServiceImpl ordering_service(
        wsv,
        max_proposal,
        rxcpp::observable<>::interval(commit_delay,
                                      rxcpp::observe_on_new_thread()),
        fake_transport,
        fake_persistent_state,
        true);

    auto on_tx = [&]() {
      for (int i = 0; i < 1000; ++i) {
        ordering_service.onTransaction(getTx());
      }
    };

    std::thread thread(on_tx);
    thread.join();
  }
  EXPECT_CALL(*fake_transport, publishProposalProxy(_, _)).Times(0);
}