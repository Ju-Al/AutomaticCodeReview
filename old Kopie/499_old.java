/*
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */
package tech.pegasys.pantheon.consensus.ibft.jsonrpc.methods;

import tech.pegasys.pantheon.consensus.ibft.IbftBlockInterface;
import tech.pegasys.pantheon.ethereum.core.BlockHeader;
import tech.pegasys.pantheon.ethereum.jsonrpc.internal.JsonRpcRequest;
import tech.pegasys.pantheon.ethereum.jsonrpc.internal.methods.AbstractBlockParameterMethod;
import tech.pegasys.pantheon.ethereum.jsonrpc.internal.methods.JsonRpcMethod;
import tech.pegasys.pantheon.ethereum.jsonrpc.internal.parameters.BlockParameter;
import tech.pegasys.pantheon.ethereum.jsonrpc.internal.parameters.JsonRpcParameter;
import tech.pegasys.pantheon.ethereum.jsonrpc.internal.queries.BlockchainQueries;

import java.util.Optional;

public class IbftGetValidatorsByBlockNumber extends AbstractBlockParameterMethod
    implements JsonRpcMethod {

  private final IbftBlockInterface ibftBlockInterface;

  public IbftGetValidatorsByBlockNumber(
      final BlockchainQueries blockchainQueries,
      final IbftBlockInterface ibftBlockInterface,
      final JsonRpcParameter parameters) {
    super(blockchainQueries, parameters);
    this.ibftBlockInterface = ibftBlockInterface;
  }

  @Override
  protected BlockParameter blockParameter(final JsonRpcRequest request) {
    return parameters().required(request.getParams(), 0, BlockParameter.class);
  }

  @Override
  protected Object resultByBlockNumber(final JsonRpcRequest request, final long blockNumber) {
    Optional<BlockHeader> blockHeader = blockchainQueries().getBlockHeaderByNumber(blockNumber);
    return blockHeader.map(header -> ibftBlockInterface.validatorsInBlock(header)).orElse(null);
  }

  @Override
  public String getName() {
    return "ibft_getValidatorsByBlockNumber";
  }
}
