                }
              }),
          std::move(factory));
    }

    auto OnDemandOrderingInit::createService(size_t max_size) {
      return std::make_shared<ordering::OnDemandOrderingServiceImpl>(max_size);
    }

    std::shared_ptr<iroha::network::OrderingGate>
    OnDemandOrderingInit::initOrderingGate(
        size_t max_size,
        std::chrono::milliseconds delay,
        std::vector<shared_model::interface::types::HashType> hashes,
        std::shared_ptr<ametsuchi::PeerQueryFactory> peer_query_factory,
        std::shared_ptr<
            ordering::transport::OnDemandOsServerGrpc::TransportFactoryType>
            transaction_factory,
        std::shared_ptr<shared_model::interface::TransactionBatchParser>
            batch_parser,
        std::shared_ptr<shared_model::interface::TransactionBatchFactory>
            transaction_batch_factory,
        std::shared_ptr<network::AsyncGrpcClient<google::protobuf::Empty>>
            async_call,
        std::unique_ptr<shared_model::interface::UnsafeProposalFactory>
            factory) {
      auto ordering_service = createService(max_size);
      service = std::make_shared<ordering::transport::OnDemandOsServerGrpc>(
          ordering_service,
          std::move(transaction_factory),
          std::move(batch_parser),
          std::move(transaction_batch_factory));
      return createGate(ordering_service,
                        createConnectionManager(std::move(peer_query_factory),
                                                std::move(async_call),
                                                delay,
                                                std::move(hashes)),
                        std::move(factory));
    }

  }  // namespace network
}  // namespace iroha
