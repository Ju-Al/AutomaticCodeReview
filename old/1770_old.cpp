/**
              return error_response<
                  shared_model::interface::NoAssetErrorResponse>;
            return apply(range.front(), [&q](auto &domain_id, auto &precision) {
              return QueryResponseBuilder().assetResponse(
                  q.assetId(), domain_id, precision);
            });
          });
 * Copyright Soramitsu Co., Ltd. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

#include "ametsuchi/impl/postgres_query_executor.hpp"

#include <boost-tuple.h>
#include <soci/boost-tuple.h>
#include <soci/postgresql/soci-postgresql.h>
#include <boost/algorithm/string.hpp>
#include <boost/format.hpp>
#include <boost/range/adaptor/filtered.hpp>
#include <boost/range/adaptor/transformed.hpp>
#include <boost/range/algorithm/for_each.hpp>
#include <boost/range/algorithm/transform.hpp>
#include <boost/range/irange.hpp>
#include "ametsuchi/impl/soci_utils.hpp"
#include "backend/protobuf/permissions.hpp"
#include "common/types.hpp"
#include "cryptography/public_key.hpp"
#include "interfaces/queries/blocks_query.hpp"

using namespace shared_model::interface::permissions;

namespace {

  using namespace iroha;

  shared_model::interface::types::DomainIdType getDomainFromName(
      const shared_model::interface::types::AccountIdType &account_id) {
    // TODO 03.10.18 andrei: IR-1728 Move getDomainFromName to shared_model
    std::vector<std::string> res;
    boost::split(res, account_id, boost::is_any_of("@"));
    return res.at(1);
  }

  std::string checkAccountRolePermission(
      shared_model::interface::permissions::Role permission,
      const std::string &account_alias = "role_account_id") {
    const auto perm_str =
        shared_model::interface::RolePermissionSet({permission}).toBitstring();
    const auto bits = shared_model::interface::RolePermissionSet::size();
    // TODO 14.09.18 andrei: IR-1708 Load SQL from separate files
    std::string query = (boost::format(R"(
          SELECT (COALESCE(bit_or(rp.permission), '0'::bit(%1%))
          & '%2%') = '%2%' AS perm FROM role_has_permissions AS rp
              JOIN account_has_roles AS ar on ar.role_id = rp.role_id
              WHERE ar.account_id = :%3%)")
                         % bits % perm_str % account_alias)
                            .str();
    return query;
  }

  /**
   * Generate an SQL subquery which checks if creator has corresponding
   * permissions for target account
   * It verifies individual, domain, and global permissions, and returns true if
   * any of listed permissions is present
   */
  auto hasQueryPermission(
      const shared_model::interface::types::AccountIdType &creator,
      const shared_model::interface::types::AccountIdType &target_account,
      Role indiv_permission_id,
      Role all_permission_id,
      Role domain_permission_id) {
    const auto bits = shared_model::interface::RolePermissionSet::size();
    const auto perm_str =
        shared_model::interface::RolePermissionSet({indiv_permission_id})
            .toBitstring();
    const auto all_perm_str =
        shared_model::interface::RolePermissionSet({all_permission_id})
            .toBitstring();
    const auto domain_perm_str =
        shared_model::interface::RolePermissionSet({domain_permission_id})
            .toBitstring();

    boost::format cmd(R"(
    WITH
        has_indiv_perm AS (
          SELECT (COALESCE(bit_or(rp.permission), '0'::bit(%1%))
          & '%3%') = '%3%' FROM role_has_permissions AS rp
              JOIN account_has_roles AS ar on ar.role_id = rp.role_id
              WHERE ar.account_id = '%2%'
        ),
        has_all_perm AS (
          SELECT (COALESCE(bit_or(rp.permission), '0'::bit(%1%))
          & '%4%') = '%4%' FROM role_has_permissions AS rp
              JOIN account_has_roles AS ar on ar.role_id = rp.role_id
              WHERE ar.account_id = '%2%'
        ),
        has_domain_perm AS (
          SELECT (COALESCE(bit_or(rp.permission), '0'::bit(%1%))
          & '%5%') = '%5%' FROM role_has_permissions AS rp
              JOIN account_has_roles AS ar on ar.role_id = rp.role_id
              WHERE ar.account_id = '%2%'
        )
    SELECT ('%2%' = '%6%' AND (SELECT * FROM has_indiv_perm))
        OR (SELECT * FROM has_all_perm)
        OR ('%7%' = '%8%' AND (SELECT * FROM has_domain_perm)) AS perm
    )");

    return (cmd % bits % creator % perm_str % all_perm_str % domain_perm_str
            % target_account % getDomainFromName(creator)
            % getDomainFromName(target_account))
        .str();
  }

  /// Query result is a tuple of optionals, since there could be no entry
  template <typename... Value>
  using QueryType = boost::tuple<boost::optional<Value>...>;

  /**
   * Create an error response in case user does not have permissions to perform
   * a query
   * @tparam Roles - type of roles
   * @param roles, which user lacks
   * @return lambda returning the error response itself
   */
  template <typename... Roles>
  auto notEnoughPermissionsResponse(Roles... roles) {
    return [roles...] {
      std::string error = "user must have at least one of the permissions: ";
      for (auto role : {roles...}) {
        error += shared_model::proto::permissions::toString(role) + ", ";
      }
      return error;
    };
  }

}  // namespace

namespace iroha {
  namespace ametsuchi {

    template <typename RangeGen, typename Pred>
    std::vector<std::unique_ptr<shared_model::interface::Transaction>>
    PostgresQueryExecutorVisitor::getTransactionsFromBlock(uint64_t block_id,
                                                           RangeGen &&range_gen,
                                                           Pred &&pred) {
      std::vector<std::unique_ptr<shared_model::interface::Transaction>>
          result{};
      auto serialized_block = block_store_.get(block_id);
      if (not serialized_block) {
        log_->error("Failed to retrieve block with id {}", block_id);
        return result;
      }
      auto deserialized_block =
          converter_->deserialize(bytesToString(*serialized_block));
      // boost::get of pointer returns pointer to requested type, or nullptr
      if (auto e =
              boost::get<expected::Error<std::string>>(&deserialized_block)) {
        log_->error(e->error);
        return result;
      }

      auto &block =
          boost::get<
              expected::Value<std::unique_ptr<shared_model::interface::Block>>>(
              deserialized_block)
              .value;

      boost::transform(range_gen(boost::size(block->transactions()))
                           | boost::adaptors::transformed(
                                 [&block](auto i) -> decltype(auto) {
                                   return block->transactions()[i];
                                 })
                           | boost::adaptors::filtered(pred),
                       std::back_inserter(result),
                       [&](const auto &tx) { return clone(tx); });

      return result;
    }

    template <typename Q,
              typename P,
              typename F,
              typename B,
              typename ErrResponse>
    QueryExecutorResult PostgresQueryExecutorVisitor::executeQuery(
        F &&f, B &&b, ErrResponse &&err_response) {
      using T = concat<Q, P>;
      try {
        soci::rowset<T> st = std::forward<F>(f)();
        auto range = boost::make_iterator_range(st.begin(), st.end());

        return apply(
            viewPermissions<P>(range.front()),
            [this, range, &b, err_response = std::move(err_response)](
                auto... perms) {
              bool temp[] = {not perms...};
              if (std::all_of(std::begin(temp), std::end(temp), [](auto b) {
                    return b;
                  })) {
                return query_response_factory_->createErrorQueryResponse(
                    QueryErrorType::kStatefulFailed,
                    err_response(),
                    this->query_hash_);
              }
              auto query_range = range
                  | boost::adaptors::transformed([](auto &t) {
                                   return rebind(viewQuery<Q>(t));
                                 })
                  | boost::adaptors::filtered([](const auto &t) {
                                   return static_cast<bool>(t);
                                 })
                  | boost::adaptors::transformed([](auto t) { return *t; });
              return std::forward<B>(b)(query_range, perms...);
            });
      } catch (const std::exception &e) {
        log_->error("Failed to execute query: {}", e.what());
        return query_response_factory_->createErrorQueryResponse(
            QueryErrorType::kStatefulFailed,
            "failed to execute query: " + std::string(e.what()),
            query_hash_);
      }
    }

    PostgresQueryExecutor::PostgresQueryExecutor(
        std::unique_ptr<soci::session> sql,
        std::shared_ptr<shared_model::interface::CommonObjectsFactory> factory,
        KeyValueStorage &block_store,
        std::shared_ptr<PendingTransactionStorage> pending_txs_storage,
        std::shared_ptr<shared_model::interface::BlockJsonConverter> converter,
        std::shared_ptr<shared_model::interface::QueryResponseFactory>
            response_factory)
        : sql_(std::move(sql)),
          block_store_(block_store),
          factory_(std::move(factory)),
          pending_txs_storage_(std::move(pending_txs_storage)),
          visitor_(*sql_,
                   factory_,
                   block_store_,
                   pending_txs_storage_,
                   std::move(converter),
                   response_factory),
          query_response_factory_{std::move(response_factory)},
          log_(logger::log("PostgresQueryExecutor")) {}

    QueryExecutorResult PostgresQueryExecutor::validateAndExecute(
        const shared_model::interface::Query &query) {
      visitor_.setCreatorId(query.creatorAccountId());
      visitor_.setQueryHash(query.hash());
      return boost::apply_visitor(visitor_, query.get());
    }

    bool PostgresQueryExecutor::validate(
        const shared_model::interface::BlocksQuery &query) {
      using T = boost::tuple<int>;
      boost::format cmd(R"(%s)");
      try {
        soci::rowset<T> st =
            (sql_->prepare
                 << (cmd % checkAccountRolePermission(Role::kGetBlocks)).str(),
             soci::use(query.creatorAccountId(), "role_account_id"));

        return st.begin()->get<0>();
      } catch (const std::exception &e) {
        log_->error("Failed to validate query: {}", e.what());
        return false;
      }
    }

    PostgresQueryExecutorVisitor::PostgresQueryExecutorVisitor(
        soci::session &sql,
        std::shared_ptr<shared_model::interface::CommonObjectsFactory> factory,
        KeyValueStorage &block_store,
        std::shared_ptr<PendingTransactionStorage> pending_txs_storage,
        std::shared_ptr<shared_model::interface::BlockJsonConverter> converter,
        std::shared_ptr<shared_model::interface::QueryResponseFactory>
            response_factory)
        : sql_(sql),
          block_store_(block_store),
          common_objects_factory_(std::move(factory)),
          pending_txs_storage_(std::move(pending_txs_storage)),
          converter_(std::move(converter)),
          query_response_factory_{std::move(response_factory)},
          log_(logger::log("PostgresQueryExecutorVisitor")) {}

    void PostgresQueryExecutorVisitor::setCreatorId(
        const shared_model::interface::types::AccountIdType &creator_id) {
      creator_id_ = creator_id;
    }

    void PostgresQueryExecutorVisitor::setQueryHash(
        const shared_model::crypto::Hash &query_hash) {
      query_hash_ = query_hash;
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetAccount &q) {
      using Q = QueryType<shared_model::interface::types::AccountIdType,
                          shared_model::interface::types::DomainIdType,
                          shared_model::interface::types::QuorumType,
                          shared_model::interface::types::DetailType,
                          std::string>;
      using P = boost::tuple<int>;

      auto cmd = (boost::format(R"(WITH has_perms AS (%s),
      t AS (
          SELECT a.account_id, a.domain_id, a.quorum, a.data, ARRAY_AGG(ar.role_id) AS roles
          FROM account AS a, account_has_roles AS ar
          WHERE a.account_id = :target_account_id
          AND ar.account_id = a.account_id
          GROUP BY a.account_id
      )
      SELECT account_id, domain_id, quorum, data, roles, perm
      FROM t RIGHT OUTER JOIN has_perms AS p ON TRUE
      )")
                  % hasQueryPermission(creator_id_,
                                       q.accountId(),
                                       Role::kGetMyAccount,
                                       Role::kGetAllAccounts,
                                       Role::kGetDomainAccounts))
                     .str();

      auto query_apply = [this](auto &account_id,
                                auto &domain_id,
                                auto &quorum,
                                auto &data,
                                auto &roles_str) {
        // TODO [IR-1750] Akvinikym 10.10.18: Make QueryResponseFactory accept
        // parameters for objects creation
        return common_objects_factory_
            ->createAccount(account_id, domain_id, quorum, data)
            .match(
                [this, roles_str = roles_str.substr(1, roles_str.size() - 2)](
                    auto &v) {
                  std::vector<shared_model::interface::types::RoleIdType> roles;

                  boost::split(
                      roles, roles_str, [](char c) { return c == ','; });

                  return query_response_factory_->createAccountResponse(
                      std::move(v.value), std::move(roles), query_hash_);
                },
                [this](expected::Error<std::string> &e) {
                  log_->error(e.error);
                  return query_response_factory_->createErrorQueryResponse(
                      QueryErrorType::kStatefulFailed,
                      "could not create account object: " + e.error,
                      query_hash_);
                });
      };

      return executeQuery<Q, P>(
          [&] {
            return (sql_.prepare << cmd,
                    soci::use(q.accountId(), "target_account_id"));
          },
          [&](auto range, auto &) {
            if (range.empty()) {
              return query_response_factory_->createErrorQueryResponse(
                  QueryErrorType::kNoAccount,
                  "no account with such id found: " + q.accountId(),
                  query_hash_);
            }

            return apply(range.front(), query_apply);
          },
          notEnoughPermissionsResponse(Role::kGetMyAccount,
                                       Role::kGetAllAccounts,
                                       Role::kGetDomainAccounts));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetSignatories &q) {
      using Q = QueryType<std::string>;
      using P = boost::tuple<int>;

      auto cmd = (boost::format(R"(WITH has_perms AS (%s),
      t AS (
          SELECT public_key FROM account_has_signatory
          WHERE account_id = :account_id
      )
      SELECT public_key, perm FROM t
      RIGHT OUTER JOIN has_perms ON TRUE
      )")
                  % hasQueryPermission(creator_id_,
                                       q.accountId(),
                                       Role::kGetMySignatories,
                                       Role::kGetAllSignatories,
                                       Role::kGetDomainSignatories))
                     .str();

      return executeQuery<Q, P>(
          [&] { return (sql_.prepare << cmd, soci::use(q.accountId())); },
          [&](auto range, auto &) {
            if (range.empty()) {
              return query_response_factory_->createErrorQueryResponse(
                  QueryErrorType::kNoSignatories,
                  "no signatories found in account with such id: "
                      + q.accountId(),
                  query_hash_);
            }

            auto pubkeys = boost::copy_range<
                std::vector<shared_model::interface::types::PubkeyType>>(
                range | boost::adaptors::transformed([](auto t) {
                  return apply(t, [&](auto &public_key) {
                    return shared_model::interface::types::PubkeyType{
                        shared_model::crypto::Blob::fromHexString(public_key)};
                  });
                }));

            return query_response_factory_->createSignatoriesResponse(
                pubkeys, query_hash_);
          },
          notEnoughPermissionsResponse(Role::kGetMySignatories,
                                       Role::kGetAllSignatories,
                                       Role::kGetDomainSignatories));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetAccountTransactions &q) {
      using Q = QueryType<shared_model::interface::types::HeightType, uint64_t>;
      using P = boost::tuple<int>;

      auto cmd = (boost::format(R"(WITH has_perms AS (%s),
      t AS (
          SELECT DISTINCT has.height, index
          FROM height_by_account_set AS has
          JOIN index_by_creator_height AS ich ON has.height = ich.height
          AND has.account_id = ich.creator_id
          WHERE account_id = :account_id
          ORDER BY has.height, index ASC
      )
      SELECT height, index, perm FROM t
      RIGHT OUTER JOIN has_perms ON TRUE
      )")
                  % hasQueryPermission(creator_id_,
                                       q.accountId(),
                                       Role::kGetMyAccTxs,
                                       Role::kGetAllAccTxs,
                                       Role::kGetDomainAccTxs))
                     .str();

      return executeQuery<Q, P>(
          [&] { return (sql_.prepare << cmd, soci::use(q.accountId())); },
          [&](auto range, auto &) {
            std::map<uint64_t, std::vector<uint64_t>> index;
            boost::for_each(range, [&index](auto t) {
              apply(t, [&index](auto &height, auto &idx) {
                index[height].push_back(idx);
              });
            });

            std::vector<std::unique_ptr<shared_model::interface::Transaction>>
                response_txs;
            for (auto &block : index) {
              auto txs = this->getTransactionsFromBlock(
                  block.first,
                  [&block](auto) { return block.second; },
                  [](auto &) { return true; });
              std::move(
                  txs.begin(), txs.end(), std::back_inserter(response_txs));
            }

            return query_response_factory_->createTransactionsResponse(
                std::move(response_txs), query_hash_);
          },
          notEnoughPermissionsResponse(
              Role::kGetMyAccTxs, Role::kGetAllAccTxs, Role::kGetDomainAccTxs));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetTransactions &q) {
      auto escape = [](auto &hash) { return "'" + hash.hex() + "'"; };
      std::string hash_str = std::accumulate(
          std::next(q.transactionHashes().begin()),
          q.transactionHashes().end(),
          escape(q.transactionHashes().front()),
          [&escape](auto &acc, auto &val) { return acc + "," + escape(val); });

      using Q =
          QueryType<shared_model::interface::types::HeightType, std::string>;
      using P = boost::tuple<int, int>;

      auto cmd = (boost::format(R"(WITH has_my_perm AS (%s),
      has_all_perm AS (%s),
      t AS (
          SELECT height, hash FROM height_by_hash WHERE hash IN (%s)
      )
      SELECT height, hash, has_my_perm.perm, has_all_perm.perm FROM t
      RIGHT OUTER JOIN has_my_perm ON TRUE
      RIGHT OUTER JOIN has_all_perm ON TRUE
      )") % checkAccountRolePermission(Role::kGetMyTxs, "account_id")
                  % checkAccountRolePermission(Role::kGetAllTxs, "account_id")
                  % hash_str)
                     .str();

      return executeQuery<Q, P>(
          [&] {
            return (sql_.prepare << cmd, soci::use(creator_id_, "account_id"));
          },
          [&](auto range, auto &my_perm, auto &all_perm) {
            std::map<uint64_t, std::unordered_set<std::string>> index;
            boost::for_each(range, [&index](auto t) {
              apply(t, [&index](auto &height, auto &hash) {
                index[height].insert(hash);
              });
            });

            std::vector<std::unique_ptr<shared_model::interface::Transaction>>
                response_txs;
            for (auto &block : index) {
              auto txs = this->getTransactionsFromBlock(
                  block.first,
                  [](auto size) {
                    return boost::irange(static_cast<decltype(size)>(0), size);
                  },
                  [&](auto &tx) {
                    return block.second.count(tx.hash().hex()) > 0
                        and (all_perm
                             or (my_perm
                                 and tx.creatorAccountId() == creator_id_));
                  });
              std::move(
                  txs.begin(), txs.end(), std::back_inserter(response_txs));
            }

            return query_response_factory_->createTransactionsResponse(
                std::move(response_txs), query_hash_);
          },
          notEnoughPermissionsResponse(Role::kGetMyTxs, Role::kGetAllTxs));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetAccountAssetTransactions &q) {
      using Q = QueryType<shared_model::interface::types::HeightType, uint64_t>;
      using P = boost::tuple<int>;

      auto cmd = (boost::format(R"(WITH has_perms AS (%s),
      t AS (
          SELECT DISTINCT has.height, index
          FROM height_by_account_set AS has
          JOIN index_by_id_height_asset AS ich ON has.height = ich.height
          AND has.account_id = ich.id
          WHERE account_id = :account_id
          AND asset_id = :asset_id
          ORDER BY has.height, index ASC
      )
      SELECT height, index, perm FROM t
      RIGHT OUTER JOIN has_perms ON TRUE
      )")
                  % hasQueryPermission(creator_id_,
                                       q.accountId(),
                                       Role::kGetMyAccAstTxs,
                                       Role::kGetAllAccAstTxs,
                                       Role::kGetDomainAccAstTxs))
                     .str();

      return executeQuery<Q, P>(
          [&] {
            return (sql_.prepare << cmd,
                    soci::use(q.accountId(), "account_id"),
                    soci::use(q.assetId(), "asset_id"));
          },
          [&](auto range, auto &) {
            std::map<uint64_t, std::vector<uint64_t>> index;
            boost::for_each(range, [&index](auto t) {
              apply(t, [&index](auto &height, auto &idx) {
                index[height].push_back(idx);
              });
            });

            std::vector<std::unique_ptr<shared_model::interface::Transaction>>
                response_txs;
            for (auto &block : index) {
              auto txs = this->getTransactionsFromBlock(
                  block.first,
                  [&block](auto) { return block.second; },
                  [](auto &) { return true; });
              std::move(
                  txs.begin(), txs.end(), std::back_inserter(response_txs));
            }

            return query_response_factory_->createTransactionsResponse(
                std::move(response_txs), query_hash_);
          },
          notEnoughPermissionsResponse(Role::kGetMyAccAstTxs,
                                       Role::kGetAllAccAstTxs,
                                       Role::kGetDomainAccAstTxs));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetAccountAssets &q) {
      using Q = QueryType<shared_model::interface::types::AccountIdType,
                          shared_model::interface::types::AssetIdType,
                          std::string>;
      using P = boost::tuple<int>;

      auto cmd = (boost::format(R"(WITH has_perms AS (%s),
      t AS (
          SELECT * FROM account_has_asset
          WHERE account_id = :account_id
      )
      SELECT account_id, asset_id, amount, perm FROM t
      RIGHT OUTER JOIN has_perms ON TRUE
      )")
                  % hasQueryPermission(creator_id_,
                                       q.accountId(),
                                       Role::kGetMyAccAst,
                                       Role::kGetAllAccAst,
                                       Role::kGetDomainAccAst))
                     .str();

      return executeQuery<Q, P>(
          [&] { return (sql_.prepare << cmd, soci::use(q.accountId())); },
          [&](auto range, auto &) {
            std::vector<std::unique_ptr<shared_model::interface::AccountAsset>>
                account_assets;
            boost::for_each(range, [this, &account_assets](auto t) {
              apply(t,
                    [this, &account_assets](
                        auto &account_id, auto &asset_id, auto &amount) {
                      common_objects_factory_
                          ->createAccountAsset(
                              account_id,
                              asset_id,
                              shared_model::interface::Amount(amount))
                          .match(
                              [&account_assets](auto &v) {
                                account_assets.push_back(clone(*v.value.get()));
                              },
                              [this](expected::Error<std::string> &e) {
                                log_->error(
                                    "could not create account asset object: {}",
                                    e.error);
                              });
                    });
            });

            return query_response_factory_->createAccountAssetResponse(
                std::move(account_assets), query_hash_);
          },
          notEnoughPermissionsResponse(
              Role::kGetMyAccAst, Role::kGetAllAccAst, Role::kGetDomainAccAst));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetAccountDetail &q) {
      using Q = QueryType<shared_model::interface::types::DetailType>;
      using P = boost::tuple<int>;

      std::string query_detail;
      if (q.key() and q.writer()) {
        auto filled_json = (boost::format("{\"%s\", \"%s\"}") % q.writer().get()
                            % q.key().get());
        query_detail = (boost::format(R"(SELECT json_build_object('%s'::text,
            json_build_object('%s'::text, (SELECT data #>> '%s'
            FROM account WHERE account_id = :account_id))) AS json)")
                        % q.writer().get() % q.key().get() % filled_json)
                           .str();
      } else if (q.key() and not q.writer()) {
        query_detail =
            (boost::format(
                 R"(SELECT json_object_agg(key, value) AS json FROM (SELECT
            json_build_object(kv.key, json_build_object('%1%'::text,
            kv.value -> '%1%')) FROM jsonb_each((SELECT data FROM account
            WHERE account_id = :account_id)) kv WHERE kv.value ? '%1%') AS
            jsons, json_each(json_build_object))")
             % q.key().get())
                .str();
      } else if (not q.key() and q.writer()) {
        query_detail = (boost::format(R"(SELECT json_build_object('%1%'::text,
          (SELECT data -> '%1%' FROM account WHERE account_id =
           :account_id)) AS json)")
                        % q.writer().get())
                           .str();
      } else {
        query_detail = (boost::format(R"(SELECT data#>>'{}' AS json FROM account
            WHERE account_id = :account_id)"))
                           .str();
      }
      auto cmd = (boost::format(R"(WITH has_perms AS (%s),
      detail AS (%s)
      SELECT json, perm FROM detail
      RIGHT OUTER JOIN has_perms ON TRUE
      )")
                  % hasQueryPermission(creator_id_,
                                       q.accountId(),
                                       Role::kGetMyAccDetail,
                                       Role::kGetAllAccDetail,
                                       Role::kGetDomainAccDetail)
                  % query_detail)
                     .str();

      return executeQuery<Q, P>(
          [&] {
            return (sql_.prepare << cmd,
                    soci::use(q.accountId(), "account_id"));
          },
          [&](auto range, auto &) {
            if (range.empty()) {
              return query_response_factory_->createErrorQueryResponse(
                  QueryErrorType::kNoAccountDetail,
                  "no details in account with such id: " + q.accountId(),
                  query_hash_);
            }

            return apply(range.front(), [this](auto &json) {
              return query_response_factory_->createAccountDetailResponse(
                  json, query_hash_);
            });
          },
          notEnoughPermissionsResponse(Role::kGetMyAccDetail,
                                       Role::kGetAllAccDetail,
                                       Role::kGetDomainAccDetail));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetRoles &q) {
      using Q = QueryType<shared_model::interface::types::RoleIdType>;
      using P = boost::tuple<int>;

      auto cmd = (boost::format(
                      R"(WITH has_perms AS (%s)
      SELECT role_id, perm FROM role
      RIGHT OUTER JOIN has_perms ON TRUE
      )") % checkAccountRolePermission(Role::kGetRoles))
                     .str();

      return executeQuery<Q, P>(
          [&] {
            return (sql_.prepare << cmd,
                    soci::use(creator_id_, "role_account_id"));
          },
          [&](auto range, auto &) {
            auto roles = boost::copy_range<
                std::vector<shared_model::interface::types::RoleIdType>>(
                range | boost::adaptors::transformed([](auto t) {
                  return apply(t, [](auto &role_id) { return role_id; });
                }));

            return query_response_factory_->createRolesResponse(roles,
                                                                query_hash_);
          },
          notEnoughPermissionsResponse(Role::kGetRoles));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetRolePermissions &q) {
      using Q = QueryType<std::string>;
      using P = boost::tuple<int>;

      auto cmd = (boost::format(
                      R"(WITH has_perms AS (%s),
      perms AS (SELECT permission FROM role_has_permissions
                WHERE role_id = :role_name)
      SELECT permission, perm FROM perms
      RIGHT OUTER JOIN has_perms ON TRUE
      )") % checkAccountRolePermission(Role::kGetRoles))
                     .str();

      return executeQuery<Q, P>(
          [&] {
            return (sql_.prepare << cmd,
                    soci::use(creator_id_, "role_account_id"),
                    soci::use(q.roleId(), "role_name"));
          },
          [&](auto range, auto &) {
            if (range.empty()) {
              return query_response_factory_->createErrorQueryResponse(
                  QueryErrorType::kNoRoles,
                  "no role " + q.roleId()
                      + " in account with such id: " + creator_id_,
                  query_hash_);
            }

            return apply(range.front(), [this](auto &permission) {
              return query_response_factory_->createRolePermissionsResponse(
                  shared_model::interface::RolePermissionSet(permission),
                  query_hash_);
            });
          },
          notEnoughPermissionsResponse(Role::kGetRoles));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetAssetInfo &q) {
      using Q =
          QueryType<shared_model::interface::types::DomainIdType, uint32_t>;
      using P = boost::tuple<int>;

      auto cmd = (boost::format(
                      R"(WITH has_perms AS (%s),
      perms AS (SELECT domain_id, precision FROM asset
                WHERE asset_id = :asset_id)
      SELECT domain_id, precision, perm FROM perms
      RIGHT OUTER JOIN has_perms ON TRUE
      )") % checkAccountRolePermission(Role::kReadAssets))
                     .str();

      return executeQuery<Q, P>(
          [&] {
            return (sql_.prepare << cmd,
                    soci::use(creator_id_, "role_account_id"),
                    soci::use(q.assetId(), "asset_id"));
          },
          [&](auto range, auto &) {
            if (range.empty()) {
              return query_response_factory_->createErrorQueryResponse(
                  QueryErrorType::kNoAsset,
                  "no asset " + q.assetId()
                      + " in account with such id: " + creator_id_,
                  query_hash_);
            }

            return apply(
                range.front(), [this, &q](auto &domain_id, auto &precision) {
                  auto asset = common_objects_factory_->createAsset(
                      q.assetId(), domain_id, precision);
                  return asset.match(
                      [this](expected::Value<std::unique_ptr<
                                 shared_model::interface::Asset>> &asset) {
                        return query_response_factory_->createAssetResponse(
                            std::move(asset.value), query_hash_);
                      },
                      [this](const expected::Error<std::string> &err) {
                        log_->error("could not create asset: {}", err.error);
                        return query_response_factory_
                            ->createErrorQueryResponse(
                                QueryErrorType::kStatefulFailed,
                                "could not create asset object: " + err.error,
                                query_hash_);
                      });
                });
          },
          notEnoughPermissionsResponse(Role::kReadAssets));
    }

    QueryExecutorResult PostgresQueryExecutorVisitor::operator()(
        const shared_model::interface::GetPendingTransactions &q) {
      std::vector<std::unique_ptr<shared_model::interface::Transaction>>
          response_txs;
      auto interface_txs =
          pending_txs_storage_->getPendingTransactions(creator_id_);
      response_txs.reserve(interface_txs.size());

      std::transform(interface_txs.begin(),
                     interface_txs.end(),
                     std::back_inserter(response_txs),
                     [](auto &tx) { return clone(*tx); });
      return query_response_factory_->createTransactionsResponse(
          std::move(response_txs), query_hash_);
    }

  }  // namespace ametsuchi
}  // namespace iroha