			"auto", -1, false)

		totalFanout := 0 // keeps summation of nodes' fanout for statistical reason

		// creates topology of the nodes
		for i, id := range ids {
			fanout := suite.topologyScenario(id.NodeID, subMngrs[i], ids, state)
			adjMap[id.NodeID] = fanout
			systemHist.Update(float64(len(fanout)))
			totalFanout += len(fanout)
		}

		if !suite.skip() {
			// prints fanout histogram of this system
			fmt.Println(systemHist.Draw())
		}

		// keeps track of average fanout per node
		aveHist.Update(float64(totalFanout) / float64(len(ids)))

		// checks end-to-end connectedness of the topology
		topology.CheckConnectedness(suite.T(), adjMap, ids)
	}

	if !suite.skip() {
		fmt.Println(aveHist.Draw())
	}
}

// topologyScenario is a test helper that creates a StatefulTopologyManager with the LinearFanoutFunc,
// it creates a TopicBasedTopology for the node and returns its fanout.
func (suite *StatefulTopologyTestSuite) topologyScenario(me flow.Identifier,
	subMngr channel.SubscriptionManager,
	ids flow.IdentityList,
	state protocol.ReadOnlyState) flow.IdentityList {
	// creates a graph sampler for the node
	graphSampler, err := topology.NewLinearFanoutGraphSampler(me)
	require.NoError(suite.T(), err)

	// creates topology of the node
	top, err := topology.NewTopicBasedTopology(me, state, graphSampler)
	require.NoError(suite.T(), err)

	// creates topology manager
	topMngr := topology.NewStatefulTopologyManager(top, subMngr)

	// generates topology of node
	myFanout, err := topMngr.MakeTopology(ids)
	require.NoError(suite.T(), err)

	return myFanout
}

// skip returns true if local environment variable AllNetworkTest is found.
func (suite *StatefulTopologyTestSuite) skip() bool {
	_, found := os.LookupEnv("AllNetworkTest")
	return found
}
