		log.Infof("Loading kube client config from path %q", kubeconfig)
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
}
