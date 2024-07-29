    using iroha::protocol::ApiBlock;
    using iroha::protocol::ApiResponse;

    iroha::protocol::ApiResponse BlockPublisherClient::sendBlock(
      const iroha::protocol::Block &block,
      const std::string &targetIp
    ){
      ApiClient client(targetIp, 50051); // TODO: Get port from config
      ApiBlock apiBlock;
      return client.receiveBlock(apiBlock);
    }


    ApiClient::ApiClient(const std::string& ip, int port) {
      // TODO(motxx): call validation of ip format and port.
      auto channel = grpc::CreateChannel(ip + ":" + std::to_string(port), grpc::InsecureChannelCredentials());
      stub_ = iroha::protocol::ApiService::NewStub(channel);
    }

    ApiResponse ApiClient::receiveBlock(const ApiBlock& block) {
      ApiResponse response;
      auto status = stub_->receiveBlock(&context_, block, &response);

      if (status.ok()) {
        return response;
      } else {
        response.Clear();
        response.set_code(iroha::protocol::FAIL);
        return response;
      }
    }

}