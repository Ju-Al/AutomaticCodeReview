      if (iroha:: instanceof
          <iroha::model::TransactionStatelessResponse>(*iroha_response)) {
        auto resp = static_cast<iroha::model::TransactionStatelessResponse &>(
            *iroha_response);
        // Find response in handler map

        auto res =
            this->handler_map_.find(resp.transaction.tx_hash.to_string());

        res->second.set_validation(
            resp.passed ? iroha::protocol::STATELESS_VALIDATION_SUCCESS
                        : iroha::protocol::STATELESS_VALIDATION_FAILED);
      }
    });
  }

  void CommandService::ToriiAsync(iroha::protocol::Transaction const &request,
                                  iroha::protocol::ToriiResponse &response) {
    auto iroha_tx = pb_factory_.deserialize(request);

    handler_map_.insert({iroha_tx.tx_hash.to_string(), response});
    // Send transaction to iroha
    tx_processor_.transaction_handle(iroha_tx);
  }

}  // namespace torii
