/*
 * Copyright 2018 ConsenSys AG.
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
package tech.pegasys.pantheon.tests.acceptance.dsl.jsonrpc;

import tech.pegasys.pantheon.tests.acceptance.dsl.condition.Condition;
import tech.pegasys.pantheon.tests.acceptance.dsl.condition.eth.ExpectEthAccountsException;
import tech.pegasys.pantheon.tests.acceptance.dsl.condition.eth.ExpectEthGetTransactionRecieptNull;
import tech.pegasys.pantheon.tests.acceptance.dsl.condition.eth.ExpectEthGetWorkException;
import tech.pegasys.pantheon.tests.acceptance.dsl.condition.eth.SanityCheckEthGetTransactionReceiptValues;
import tech.pegasys.pantheon.tests.acceptance.dsl.condition.eth.SanityCheckEthGetWorkValues;
import tech.pegasys.pantheon.tests.acceptance.dsl.transaction.eth.EthTransactions;

public class Eth {

  private final EthTransactions transactions;

  public Eth(final EthTransactions transactions) {
    this.transactions = transactions;
  }

  public Condition getWork() {
    return new SanityCheckEthGetWorkValues(transactions.getWork());
  }

  public Condition getWorkExceptional(final String expectedMessage) {
    return new ExpectEthGetWorkException(transactions.getWork(), expectedMessage);
  }

  public Condition accountsExceptional(final String expectedMessage) {
    return new ExpectEthAccountsException(transactions.accounts(), expectedMessage);
  public Condition getTransactionReceipt(final String param) {
    return new SanityCheckEthGetTransactionReceiptValues(transactions.getTransactionReceipt(param));
  }

  public Condition getTransactionReceiptNull(final String param) {
    return new ExpectEthGetTransactionRecieptNull(transactions.getTransactionReceipt(param));
  }
}
