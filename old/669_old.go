	wds []*v1alpha2.WorkloadDefinition
	tds []*v1alpha2.TraitDefinition
}

// GetWorkloadDefinition  Get WorkloadDefinition
func (mock *MockClient) GetWorkloadDefinition(name string) (*v1alpha2.WorkloadDefinition, error) {
	for _, wd := range mock.wds {
		if wd.Name == name {
			return wd, nil
		}
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{
		Group:    v1alpha2.Group,
		Resource: "WorkloadDefinition",
	}, name)
}

// GetTraitDefition Get TraitDefition
func (mock *MockClient) GetTraitDefition(name string) (*v1alpha2.TraitDefinition, error) {
	for _, td := range mock.tds {
		if td.Name == name {
			return td, nil
		}
	}
	return nil, kerrors.NewNotFound(schema.GroupResource{
		Group:    v1alpha2.Group,
		Resource: "TraitDefinition",
	}, name)
}

// AddWD  add workload definition to Mock Manager
func (mock *MockClient) AddWD(s string) error {
	wd := &v1alpha2.WorkloadDefinition{}
	_body, err := kyaml.YAMLToJSON([]byte(s))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(_body, wd); err != nil {
		return err
	}

	if mock.wds == nil {
		mock.wds = []*v1alpha2.WorkloadDefinition{}
	}
	mock.wds = append(mock.wds, wd)
	return nil
}

// AddTD add triat definition to Mock Manager
func (mock *MockClient) AddTD(s string) error {
	td := &v1alpha2.TraitDefinition{}
	_body, err := kyaml.YAMLToJSON([]byte(s))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(_body, td); err != nil {
		return err
	}
	if mock.tds == nil {
		mock.tds = []*v1alpha2.TraitDefinition{}
	}
	mock.tds = append(mock.tds, td)
	return nil
}
