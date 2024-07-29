		parts := strings.Split(line, "=")
		attrs[parts[0]] = parts[1]
	}
	return attrs, nil
}

func linuxRelease() (string, error) {
	attrs, err := linuxReleaseAttributes()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", attrs["NAME"], attrs["VERSION_ID"]), nil
}
