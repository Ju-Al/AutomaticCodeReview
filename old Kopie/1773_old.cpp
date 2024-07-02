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

#include "network/impl/block_loader_service.hpp"
#include "backend/protobuf/block.hpp"

using namespace iroha;
using namespace iroha::ametsuchi;
using namespace iroha::network;

BlockLoaderService::BlockLoaderService(
    std::shared_ptr<BlockQueryFactory> block_query_factory,
    std::shared_ptr<iroha::consensus::ConsensusResultCache>
        consensus_result_cache)
    : block_query_factory_(std::move(block_query_factory)),
      consensus_result_cache_(std::move(consensus_result_cache)),
      log_(logger::log("BlockLoaderService")) {}

grpc::Status BlockLoaderService::retrieveBlocks(
    ::grpc::ServerContext *context,
    const proto::BlocksRequest *request,
    ::grpc::ServerWriter<::iroha::protocol::Block> *writer) {
  auto blocks = block_query_factory_->createBlockQuery() |
      [height = request->height()](const auto &block_query) {
        return block_query->getBlocksFrom(height);
      };
  std::for_each(blocks.begin(), blocks.end(), [&writer](const auto &block) {
    writer->Write(std::dynamic_pointer_cast<shared_model::proto::Block>(block)
                      ->getTransport());
  });
  return grpc::Status::OK;
}

grpc::Status BlockLoaderService::retrieveBlock(
    ::grpc::ServerContext *context,
    const proto::BlockRequest *request,
    protocol::Block *response) {
  const auto hash = shared_model::crypto::Hash(request->hash());
  if (hash.size() == 0) {
    log_->error("Bad hash in request");
    return grpc::Status(grpc::StatusCode::INVALID_ARGUMENT,
                        "Bad hash provided");
  }

  if (not block) {
    log_->info("Requested to retrieve a block from an empty cache");
    return grpc::Status(grpc::StatusCode::NOT_FOUND, "Cache is empty");
  }
  if (block->hash() != hash) {
        "Requested to retrieve a block with hash other than the one in cache");
  // required block must be in the cache
  auto block = consensus_result_cache_->get();
  if (block and block->hash() == hash) {
    auto transport_block =
        std::static_pointer_cast<shared_model::proto::Block>(block)
            ->getTransport();
    response->CopyFrom(transport_block);
    return grpc::Status::OK;
  } else if (not block) {
    log_->info(
        "Tried to retrieve a block from an empty cache: requested block hash "
        "{}",
        hash.hex());
  } else {
    log_->info(
        "Requested to retrieve a block, but cache contains another block: "
        "requested {}, in cache {}",
        hash.hex(),
        block->hash().hex());
  }

  // cache missed: notify and try to fetch the block from block storage itself
  auto blocks = block_query_factory_->createBlockQuery() |
      [](const auto &block_query) { return block_query->getBlocksFrom(1); };
  auto found_block = std::find_if(
      std::begin(blocks), std::end(blocks), [&hash](const auto &block) {
        return block->hash() == hash;
      });
  if (found_block == std::end(blocks)) {
    log_->error("Could not retrieve a block from block storage: requested {}",
                hash.hex());
    return grpc::Status(grpc::StatusCode::NOT_FOUND, "Block not found");
  }

  response->CopyFrom(
      std::static_pointer_cast<shared_model::proto::Block>(*found_block)
          ->getTransport());
  return grpc::Status::OK;
}