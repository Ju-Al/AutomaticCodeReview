		for i := 0; i < switchedAttributesCount; i++ {
			if i == 0 {
				tab = []int64{1}
			} else {
				tab = append(tab, 1)
			}
		}

		for i := 0; i < switchedVectorCount; i++ {
			ciphertexts[libmedco.TempID(i)] = *libmedco.EncryptIntVector(aggregateKey, tab)
		}

		clientSecret := suite.Scalar().Pick(random.Stream)
		clientPublic := suite.Point().Mul(suite.Point().Base(), clientSecret)

		root.ProtocolInstance().(*KeySwitchingProtocol).TargetPublicKey = &clientPublic
		root.ProtocolInstance().(*KeySwitchingProtocol).TargetOfSwitch = &ciphertexts

		round := monitor.NewTimeMeasure("MEDCO_PROTOCOL")
		root.StartProtocol()
		<-root.ProtocolInstance().(*KeySwitchingProtocol).FeedbackChannel
		round.Record()

	}

	return nil
}
