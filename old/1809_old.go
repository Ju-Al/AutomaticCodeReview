		return BdnSignature(sig).Verify(testSuite, cosiProtocol.Msg, pubs)
	case <-time.After(2 * time.Second):
	}

	return errors.New("timeout")
}

// testService allows setting the pairing keys of the protocol.
type testService struct {
	// We need to embed the ServiceProcessor, so that incoming messages
	// are correctly handled.
	*onet.ServiceProcessor
}

// Starts a new service. No function needed.
func newService(c *onet.Context) (onet.Service, error) {
	s := &testService{
		ServiceProcessor: onet.NewServiceProcessor(c),
	}
	return s, nil
}
