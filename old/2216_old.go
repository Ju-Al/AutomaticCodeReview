		}
		if val == "false" || val == "no" || val == "off" || val == "" {
			return pyBool(false)
		}
		log.Fatalf("%s: Invalid boolean value %v", key, val)
	}

	if toType == "int" {
		if val == "" {
			return pyInt(0)
		}

		i, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("%s: Invalid int value %v", key, val)
		}
		return pyInt(i)
	}

	log.Fatalf("%s: invalid config type %v", key, toType)
	return pyNone{}
}
