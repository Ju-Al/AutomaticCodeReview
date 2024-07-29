	Variables map[string]string `yaml:"variables"`
	Secrets   map[string]string `yaml:"secrets"`
}

// WorkloadProps contains properties for creating a new service manifest.
type WorkloadProps struct {
	Name       string
	Dockerfile string
}

// UnmarshalWorkload deserializes the YAML input stream into a service manifest object.
// If an error occurs during deserialization, then returns the error.
// If the service type in the manifest is invalid, then returns an ErrInvalidManifestType.
func UnmarshalWorkload(in []byte) (interface{}, error) {
	am := Workload{}
	if err := yaml.Unmarshal(in, &am); err != nil {
		return nil, fmt.Errorf("unmarshal to service manifest: %w", err)
	}
	typeVal := aws.StringValue(am.Type)

	switch typeVal {
	case LoadBalancedWebServiceType:
		m := newDefaultLoadBalancedWebService()
		if err := yaml.Unmarshal(in, m); err != nil {
			return nil, fmt.Errorf("unmarshal to load balanced web service: %w", err)
		}
		return m, nil
	case BackendServiceType:
		m := newDefaultBackendService()
		if err := yaml.Unmarshal(in, m); err != nil {
			return nil, fmt.Errorf("unmarshal to backend service: %w", err)
		}
		if m.BackendServiceConfig.Image.HealthCheck != nil {
			// Make sure that unset fields in the healthcheck gets a default value.
			m.BackendServiceConfig.Image.HealthCheck.applyIfNotSet(newDefaultContainerHealthCheck())
		}
		return m, nil
	case ScheduledJobType:
		m := newDefaultScheduledJob()
		if err := yaml.Unmarshal(in, m); err != nil {
			return nil, fmt.Errorf("unmarshal to scheduled job: %w", err)
		}
		return m, nil
	default:
		return nil, &ErrInvalidWkldManifestType{Type: typeVal}
	}
}
