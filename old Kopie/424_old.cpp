/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

#include <boost/test/auto_unit_test.hpp>
#include <thrift/transport/TSocket.h>
#include <thrift/transport/TServerSocket.h>
#include "TestPortFixture.h"
#include "TestTServerSocket.h"

using apache::thrift::transport::TServerSocket;
using apache::thrift::transport::TSocket;
using apache::thrift::transport::TTransport;
using apache::thrift::transport::TTransportException;

BOOST_FIXTURE_TEST_SUITE ( TServerSocketTest, TestPortFixture )

BOOST_AUTO_TEST_CASE( test_bind_to_address )
{
    TestTServerSocket sock1("localhost", m_serverPort);
    sock1.listen();
    TSocket clientSock("localhost", m_serverPort);
    clientSock.open();
    boost::shared_ptr<TTransport> accepted = sock1.acceptImpl();
    accepted->close();
    sock1.close();

    TServerSocket sock2("this.is.truly.an.unrecognizable.address.", m_serverPort);
    TServerSocket sock2("257.258.259.260", m_serverPort);
    BOOST_CHECK_THROW(sock2.listen(), TTransportException);
    sock2.close();
}

BOOST_AUTO_TEST_CASE( test_close_before_listen )
{
    TServerSocket sock1("localhost", m_serverPort);
    sock1.close();
}

BOOST_AUTO_TEST_CASE( test_get_port )
{
    TServerSocket sock1("localHost", 888);
    BOOST_CHECK_EQUAL(888, sock1.getPort());
}

BOOST_AUTO_TEST_SUITE_END()

