/**
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
#include "model/query_execution.hpp"
#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <model/queries/responses/account_assets_response.hpp>
#include "model/queries/responses/account_response.hpp"
#include "model/queries/responses/error_response.hpp"

using ::testing::Return;
using ::testing::AtLeast;
using ::testing::_;
using ::testing::AllOf;

class WSVQueriesMock : public iroha::ametsuchi::WsvQuery {
 public:
  MOCK_METHOD1(getAccount, nonstd::optional<iroha::model::Account>(
                               const std::string &account_id));
  MOCK_METHOD1(getSignatories,
               nonstd::optional<std::vector<iroha::ed25519::pubkey_t>>(
                   const std::string &account_id));
  MOCK_METHOD1(getAsset, nonstd::optional<iroha::model::Asset>(
                             const std::string &asset_id));

  MOCK_METHOD2(getAccountAsset,
               nonstd::optional<iroha::model::AccountAsset>(
                   const std::string &account_id, const std::string &asset_id));
  MOCK_METHOD0(getPeers, nonstd::optional<std::vector<iroha::model::Peer>>());
};

class BlockQueryMock : public iroha::ametsuchi::BlockQuery {
 public:
  MOCK_METHOD1(getAccountTransactions,
               rxcpp::observable<iroha::model::Transaction>(std::string));

  MOCK_METHOD2(getBlocks,
               rxcpp::observable<iroha::model::Block>(uint32_t, uint32_t));
};

auto ACCOUNT_ID = "test@test";
auto ADMIN_ID = "admin@test";
auto DOMAIN_NAME = "test";
auto ADVERSARY_ID = "adversary@test";
auto ASSET_ID = "coin";

iroha::model::Account get_default_creator() {
  iroha::model::Account creator = iroha::model::Account();
  creator.account_id = ADMIN_ID;
  creator.domain_name = DOMAIN_NAME;
  std::fill(creator.master_key.begin(), creator.master_key.end(), 0x1);
  creator.quorum = 1;
  creator.permissions.read_all_accounts = true;
  return creator;
}

iroha::model::Account get_default_account() {
  auto dummy = iroha::model::Account();
  dummy.account_id = ACCOUNT_ID;
  dummy.domain_name = DOMAIN_NAME;
  std::fill(dummy.master_key.begin(), dummy.master_key.end(), 0x2);
  dummy.quorum = 1;
  return dummy;
}

iroha::model::Account get_default_adversary() {
  auto dummy = iroha::model::Account();
  dummy.account_id = ADVERSARY_ID;
  dummy.domain_name = DOMAIN_NAME;
  std::fill(dummy.master_key.begin(), dummy.master_key.end(), 0xF);
  dummy.quorum = 1;
  return dummy;
}

void set_default_ametsuchi(WSVQueriesMock &test_wsv,
                           BlockQueryMock &test_blocks) {
  // No account exist
  EXPECT_CALL(test_wsv, getAccount(_)).WillRepeatedly(Return(nonstd::nullopt));
  // Admin exist
  auto admin = get_default_creator();

  EXPECT_CALL(test_wsv, getAccount(ADMIN_ID)).WillRepeatedly(Return(admin));
  // Test account exist
  auto dummy = get_default_account();

  EXPECT_CALL(test_wsv, getAccount(ACCOUNT_ID)).WillRepeatedly(Return(dummy));
  auto adversary = get_default_adversary();
  EXPECT_CALL(test_wsv, getAccount(ADVERSARY_ID))
      .WillRepeatedly(Return(adversary));

  EXPECT_CALL(test_wsv, getAccountAsset(_, _))
      .WillRepeatedly(Return(nonstd::nullopt));

  auto acc_asset = iroha::model::AccountAsset();
  acc_asset.asset_id = ASSET_ID;
  acc_asset.account_id = ACCOUNT_ID;
  acc_asset.balance = 150;
  EXPECT_CALL(test_wsv, getAccountAsset(ACCOUNT_ID, ASSET_ID))
      .WillRepeatedly(Return(acc_asset));
}

TEST(QueryExecutor, get_account) {
  WSVQueriesMock wsv_queries;
  BlockQueryMock block_queries;
  auto query_proccesor =
      iroha::model::QueryProcessingFactory(wsv_queries, block_queries);

  set_default_ametsuchi(wsv_queries, block_queries);

  // Valid cases:
  // 1. Admin asks account_id
  iroha::model::GetAccount query;
  query.account_id = ACCOUNT_ID;
  query.creator_account_id = ADMIN_ID;
  query.signature.pubkey = get_default_creator().master_key;
  auto response = query_proccesor.execute(query);
  auto cast_resp =
      std::static_pointer_cast<iroha::model::AccountResponse>(response);
  ASSERT_EQ(cast_resp->account.account_id, ACCOUNT_ID);
  ASSERT_EQ(cast_resp->query.creator_account_id, ADMIN_ID);

  // 2. Account creator asks about his account
  query.account_id = ACCOUNT_ID;
  query.creator_account_id = ACCOUNT_ID;
  query.signature.pubkey = get_default_account().master_key;
  response = query_proccesor.execute(query);
  cast_resp = std::static_pointer_cast<iroha::model::AccountResponse>(response);
  ASSERT_EQ(cast_resp->account.account_id, ACCOUNT_ID);

  // --------- Non valid cases: -------

  // 1. Asking non-existed account

  query.account_id = "nonacc";
  query.creator_account_id = ADMIN_ID;
  query.signature.pubkey = get_default_creator().master_key;
  response = query_proccesor.execute(query);
  auto cast_resp_2 =
      std::dynamic_pointer_cast<iroha::model::AccountResponse>(response);
  ASSERT_EQ(cast_resp_2, nullptr);
  auto err_resp =
      std::dynamic_pointer_cast<iroha::model::ErrorResponse>(response);
  ASSERT_EQ(err_resp->reason, "No account");
  ASSERT_EQ(err_resp->query.creator_account_id, ADMIN_ID);

  // 2. No rights to ask account
  query.account_id = ACCOUNT_ID;
  query.creator_account_id = ADVERSARY_ID;
  query.signature.pubkey = get_default_adversary().master_key;
  response = query_proccesor.execute(query);
  cast_resp =
      std::dynamic_pointer_cast<iroha::model::AccountResponse>(response);
  ASSERT_EQ(cast_resp, nullptr);

  err_resp = std::dynamic_pointer_cast<iroha::model::ErrorResponse>(response);
  ASSERT_EQ(err_resp->reason, "Not valid query");

  // 3. No creator
  query.account_id = ACCOUNT_ID;
  query.creator_account_id = "noacc";
  query.signature.pubkey = get_default_adversary().master_key;
  response = query_proccesor.execute(query);
  cast_resp =
      std::dynamic_pointer_cast<iroha::model::AccountResponse>(response);
  ASSERT_EQ(cast_resp, nullptr);

  err_resp = std::dynamic_pointer_cast<iroha::model::ErrorResponse>(response);
  ASSERT_EQ(err_resp->reason, "Not valid query");

  // TODO: tests for signatures
}

TEST(QueryExecutor, get_account_asset) {
  WSVQueriesMock wsv_queries;
  BlockQueryMock block_queries;
  auto query_proccesor =
      iroha::model::QueryProcessingFactory(wsv_queries, block_queries);

  set_default_ametsuchi(wsv_queries, block_queries);

  // Valid cases:
  // 1. Admin asks account_id
  iroha::model::GetAccountAsset query;
  query.account_id = ACCOUNT_ID;
  query.asset_id = ASSET_ID;
  query.creator_account_id = ADMIN_ID;
  query.signature.pubkey = get_default_creator().master_key;
  auto response = query_proccesor.execute(query);
  auto cast_resp =
      std::static_pointer_cast<iroha::model::AccountAssetResponse>(response);
  ASSERT_EQ(cast_resp->acc_asset.account_id, ACCOUNT_ID);
  ASSERT_EQ(cast_resp->acc_asset.asset_id, ASSET_ID);
  ASSERT_EQ(cast_resp->query.creator_account_id, ADMIN_ID);

  // 2. Account creator asks about his account
  query.account_id = ACCOUNT_ID;
  query.creator_account_id = ACCOUNT_ID;
  query.signature.pubkey = get_default_account().master_key;
  response = query_proccesor.execute(query);
  cast_resp = std::static_pointer_cast<iroha::model::AccountAssetResponse>(response);
  ASSERT_EQ(cast_resp->acc_asset.account_id, ACCOUNT_ID);
  ASSERT_EQ(cast_resp->acc_asset.asset_id, ASSET_ID);
  ASSERT_EQ(cast_resp->query.creator_account_id, ACCOUNT_ID);

  // --------- Non valid cases: -------

  // 1. Asking non-existed account asset

  query.account_id = "nonacc";
  query.creator_account_id = ADMIN_ID;
  query.signature.pubkey = get_default_creator().master_key;
  response = query_proccesor.execute(query);
  auto cast_resp_2 =
      std::dynamic_pointer_cast<iroha::model::AccountAssetResponse>(response);
  ASSERT_EQ(cast_resp_2, nullptr);
  auto err_resp =
      std::dynamic_pointer_cast<iroha::model::ErrorResponse>(response);
  ASSERT_EQ(err_resp->reason, "No Account Asset");
  ASSERT_EQ(err_resp->query.creator_account_id, ADMIN_ID);


  // Asking non-existed account asset

  query.account_id = ACCOUNT_ID;
  query.asset_id = "nonasset";
  query.creator_account_id = ADMIN_ID;
  query.signature.pubkey = get_default_creator().master_key;
  response = query_proccesor.execute(query);
  cast_resp_2 =
      std::dynamic_pointer_cast<iroha::model::AccountAssetResponse>(response);
  ASSERT_EQ(cast_resp_2, nullptr);
  err_resp =
      std::dynamic_pointer_cast<iroha::model::ErrorResponse>(response);
  ASSERT_EQ(err_resp->reason, "No Account Asset");
  ASSERT_EQ(err_resp->query.creator_account_id, ADMIN_ID);

  // 2. No rights to ask
  query.account_id = ACCOUNT_ID;
  query.asset_id = ASSET_ID;
  query.creator_account_id = ADVERSARY_ID;
  query.signature.pubkey = get_default_adversary().master_key;
  response = query_proccesor.execute(query);
  cast_resp =
      std::dynamic_pointer_cast<iroha::model::AccountAssetResponse>(response);
  ASSERT_EQ(cast_resp, nullptr);

  err_resp = std::dynamic_pointer_cast<iroha::model::ErrorResponse>(response);
  ASSERT_EQ(err_resp->reason, "Not valid query");

  // 3. No creator
  query.account_id = ACCOUNT_ID;
  query.creator_account_id = "noacc";
  query.signature.pubkey = get_default_adversary().master_key;
  response = query_proccesor.execute(query);
  cast_resp =
      std::dynamic_pointer_cast<iroha::model::AccountAssetResponse>(response);
  ASSERT_EQ(cast_resp, nullptr);

  err_resp = std::dynamic_pointer_cast<iroha::model::ErrorResponse>(response);
  ASSERT_EQ(err_resp->reason, "Not valid query");

  // TODO: tests for signatures
}