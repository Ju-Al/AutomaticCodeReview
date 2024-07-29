	switch tp.Name {
	case "metrics":
		// metrics trait will rely on a Prometheus instance to be installed
		// make sure the chart is a prometheus operator
		if tp.Install == nil {
			break
		}
		if tp.Install.Helm.Namespace == "monitoring" && tp.Install.Helm.Name == "kube-prometheus-stack" {
			if err := InstallPrometheusInstance(client); err != nil {
				return err
			}
		}
	default:
	}
	return nil
}

func addSourceIntoExtension(in *runtime.RawExtension, source *types.Source) error {
	var extension map[string]interface{}
	err := json.Unmarshal(in.Raw, &extension)
	if err != nil {
		return err
	}
	extension["source"] = source
	data, err := json.Marshal(extension)
	if err != nil {
		return err
	}
	in.Raw = data
	return nil
}

// CheckLabelExistence checks whether a label `key=value` exists in definition labels
func CheckLabelExistence(labels map[string]string, label string) bool {
	splitLabel := strings.Split(label, "=")
	k, v := splitLabel[0], splitLabel[1]
	if labelValue, ok := labels[k]; ok {
		if labelValue == v {
			return true
		}
	}
	return false
}
