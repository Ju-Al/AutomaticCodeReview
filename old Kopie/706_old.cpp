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

#include "interactive/interactive_multisig.hpp"
#include <model/converters/pb_common.hpp>
#include <model/converters/pb_transaction_factory.hpp>
#include "datetime/time.hpp"
#include "grpc_response_handler.hpp"

namespace iroha_cli {
  namespace interactive {
    InteractiveMultiSigCli::InteractiveMultiSigCli(
        const std::string &account_id,
        uint64_t query_counter,
        const CliClient &client,
        std::shared_ptr<iroha::model::ModelCryptoProvider> provider)
        : client_(client),
          creator_(account_id),
          query_counter_(query_counter),
          provider_(std::move(provider)) {}

    nonstd::optional<iroha::protocol::TransactionsResponse>
    InteractiveMultiSigCli::fetchUncompletedTransactions() {
      ++query_counter_;
      auto local_time_ = iroha::time::now();
      auto query_ =
          std::make_shared<iroha::model::GetAccountUncompletedTransactions>(
              creator_);
      generator_.setQueryMetaData(
          query_, local_time_, creator_, query_counter_);

      provider_->sign(*query_);
      auto response = client_.sendQuery(query_);
      if (not response.status.ok()) {
        GrpcResponseHandler{}.handle(response);
        return nonstd::nullopt;
      }
      return response.answer.transactions_response();
    }

    void InteractiveMultiSigCli::run() {
      auto txs = fetchUncompletedTransactions();
      if (not txs.has_value()) {
        return;
      }
      // ---- Build menu with transaction ---
      MenuPoints tx_menu;
      iroha::model::converters::PbTransactionFactory pb_factory;
      for (const auto &tx : txs.value().transactions()) {
        auto iroha_tx = *pb_factory.deserialize(tx);
        uncompleted_transactions.push_back(iroha_tx);
        auto tx_hash = iroha::hash(tx);
        auto decr = "creator " + iroha_tx.creator_account_id + " hash "
            + tx_hash.to_hexstring() + " sigs/quorum "
            + std::to_string(iroha_tx.signatures.size()) + "/"
            + std::to_string(iroha_tx.quorum);
        addMenuPoint(tx_menu, decr, "");
      }
      addBackOption(tx_menu);
      // ---- Parsing pipeline ----
      while (true) {
        printMenu("Choose transaction to sign and send to iroha peer", tx_menu);
        // Read user response
        auto line = promtString("> ");
        if (isBackOption(line)) {
          break;
        }
        auto val = parser::parseValue<uint32_t>(line);
        if (not val.has_value()
            or val.value() > uncompleted_transactions.size()) {
          handleEmptyCommand();
          continue;
        }
        auto tx = uncompleted_transactions[val.value() - 1];
        // sign transaction
        provider_->sign(tx);
        // Send transaction to iroha peer
        GrpcResponseHandler{}.handle(client_.sendTx(tx));
        printEnd();
        break;
      }
    }
  }
}
