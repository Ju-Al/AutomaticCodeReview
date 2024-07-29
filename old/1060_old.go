		"/kiali/",
		"kiali/",
		"/^kiali",
	}

	for _, webroot := range validWebRoots {
		conf.Server.WebRoot = webroot
		config.Set(conf)
		if err := validateConfig(); err != nil {
			t.Errorf("Web root validation should have succeeded for [%v]: %v", conf.Server.WebRoot, err)
		}
	}

	for _, webroot := range invalidWebRoots {
		conf.Server.WebRoot = webroot
		config.Set(conf)
		if err := validateConfig(); err == nil {
			t.Errorf("Web root validation should have failed [%v]", conf.Server.WebRoot)
		}
	}
}
