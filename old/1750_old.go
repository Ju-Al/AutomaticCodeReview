		return nil, err
	}

	return svc.ds.DeleteGlobalPolicies(ids)
}
