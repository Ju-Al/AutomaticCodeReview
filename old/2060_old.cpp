/**
 * Copyright Soramitsu Co., Ltd. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

#include "module/irohad/consensus/yac/yac_mocks.hpp"

#include <grpc++/grpc++.h>

#include "consensus/yac/transport/impl/network_impl.hpp"
#include "consensus/yac/transport/yac_pb_converters.hpp"
#include "framework/mock_stream.h"
#include "network/impl/grpc_channel_builder.hpp"
#include "yac_mock.grpc.pb.h"

using ::testing::_;
using ::testing::DoAll;
using ::testing::InvokeWithoutArgs;
using ::testing::Return;
using ::testing::SaveArg;

namespace iroha {
  namespace consensus {
    namespace yac {
      class YacNetworkTest : public ::testing::Test {
       public:
        static constexpr auto default_ip = "0.0.0.0";
        static constexpr auto default_address = "0.0.0.0:0";
        void SetUp() override {
          notifications = std::make_shared<MockYacNetworkNotifications>();
          async_call = std::make_shared<
              network::AsyncGrpcClient<google::protobuf::Empty>>();
          // Custom deleter is necessary since it is necessary to construct
          // unique_ptr in client_creator in there is not possible to make empty
          // destructor there
          stub =
              std::unique_ptr<iroha::consensus::yac::proto::MockYacStub,
                              void (*)(
                                  iroha::consensus::yac::proto::MockYacStub *)>(
                  new iroha::consensus::yac::proto::MockYacStub(),
                  [](iroha::consensus::yac::proto::MockYacStub *ptr) {});
          std::function<std::unique_ptr<proto::Yac::StubInterface>(
              const shared_model::interface::Peer &)>
              client_creator([this](const shared_model::interface::Peer &peer) {
                // stub from the YacNetworkTest will be destroyed in destructor
                // of NetworkImpl
                return std::unique_ptr<
                    iroha::consensus::yac::proto::MockYacStub>(stub.get());
              });
          network = std::make_shared<NetworkImpl>(async_call, client_creator);

          message.hash.vote_hashes.proposal_hash = "proposal";
          message.hash.vote_hashes.block_hash = "block";

          // getTransport is not used in network at the moment, please check if
          // test fails
          message.hash.block_signature = createSig("");
          message.signature = createSig("");
          network->subscribe(notifications);

          int port = 0;
          peer = mk_peer(std::string(default_ip) + ":" + std::to_string(port));
        }

        std::shared_ptr<MockYacNetworkNotifications> notifications;
        std::shared_ptr<network::AsyncGrpcClient<google::protobuf::Empty>>
            async_call;
        std::shared_ptr<NetworkImpl> network;
        std::shared_ptr<shared_model::interface::Peer> peer;
        VoteMessage message;
        shared_model::crypto::PublicKey pubkey =
            shared_model::crypto::PublicKey{""};
        std::shared_ptr<iroha::consensus::yac::proto::MockYacStub> stub;
      };

      /**
       * @given initialized network
       * @when send vote to itself
       * @then vote handled
       */
      TEST_F(YacNetworkTest, MessageHandledWhenMessageSent) {
        proto::State request;
        auto r = std::make_unique<grpc::testing::MockClientAsyncResponseReader<
            google::protobuf::Empty>>();
        EXPECT_CALL(*stub, AsyncSendStateRaw(_, _, _))
            .WillOnce(DoAll(SaveArg<1>(&request), Return(r.get())));

        network->sendState(*peer, {message});

        ASSERT_EQ(request.votes_size(), 1);
      }

      /**
       * @given initialized network
       * @when send request with one vote
       * @then status OK
       */
      TEST_F(YacNetworkTest, SendMessage) {
        proto::State request;
        grpc::ServerContext context;

        for (const auto &vote : {message}) {
          auto pb_vote = request.add_votes();
          *pb_vote = PbConverters::serializeVote(vote);
        }

        auto response = network->SendState(&context, &request, nullptr);
        ASSERT_EQ(response.error_code(), grpc::StatusCode::OK);
      }

      /**
       * @given initialized network
       * @when send request with no votes
       * @then status CANCELLED
       */
      TEST_F(YacNetworkTest, SendMessageEmptyKeys) {
        proto::State request;
        grpc::ServerContext context;
        auto response = network->SendState(&context, &request, nullptr);
        ASSERT_EQ(response.error_code(), grpc::StatusCode::CANCELLED);
      }
    }  // namespace yac
  }    // namespace consensus
}  // namespace iroha
