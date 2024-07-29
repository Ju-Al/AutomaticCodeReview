  void load() {
    Properties properties = new Properties();
    try {
      File file = Paths.get(credentialsFile).toFile();
      if (!file.getName().endsWith(".properties")) {
        throw new FileNotFoundException("The file does not exist or not end with '.properties'");
      }
      try (FileInputStream is = new FileInputStream(file)) {
        properties.load(is);
        if (!properties.containsKey(USERNAME_PROP) || !properties.containsKey(PASSWORD_PROP)) {
          return;
        }
        basicCredentials.updateCredentials(
          properties.getProperty(USERNAME_PROP),
          properties.getProperty(PASSWORD_PROP)
        );
      }
    } catch (Exception e) {
      if (rateLimiter.tryAcquire()) {
        LOGGER.error("Load credentials file error", e);
      }
    }
  }
}
