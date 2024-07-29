	in := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	out, err := client.GetSecretValue(in)
	if err != nil {
		return docker.AuthConfiguration{}, errors.Wrapf(err,
			"asm fetching secret from the service for %s", secretID)
	}

	return extractASMValue(out)
}

func extractASMValue(out *secretsmanager.GetSecretValueOutput) (docker.AuthConfiguration, error) {
	if out == nil {
		return docker.AuthConfiguration{}, errors.Errorf(
			"asm fetching authorization data: empty response")
	}

	secretValue := aws.StringValue(out.SecretString)
	if secretValue == "" {
		return docker.AuthConfiguration{}, errors.Errorf(
			"asm fetching authorization data: empty secrets value")
	}

	authDataValue := AuthDataValue{}
	err := json.Unmarshal([]byte(secretValue), &authDataValue)
	if err != nil {
		// could  not unmarshal, incorrect secret value schema
		return docker.AuthConfiguration{}, errors.Wrapf(err,
			"asm fetching authorization data: unable to unmarshal secret value")
	}

	username := aws.StringValue(authDataValue.Username)
	password := aws.StringValue(authDataValue.Password)

	if username == "" {
		return docker.AuthConfiguration{}, errors.Errorf(
			"asm fetching username: AuthorizationData is malformed, empty field")
	}

	if password == "" {
		return docker.AuthConfiguration{}, errors.Errorf(
			"asm fetching password: AuthorizationData is malformed, empty field")
	}

	dac := docker.AuthConfiguration{
		Username: username,
		Password: password,
	}

	return dac, nil
}
