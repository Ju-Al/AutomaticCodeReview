    assertEquals("0% free", element.getText());
  }

  private static Server<?> createHub() {
    int publish = PortProber.findFreePort();
    int subscribe = PortProber.findFreePort();

    String[] rawConfig = new String[] {
      "[events]",
      "publish = \"tcp://localhost:" + publish + "\"",
      "subscribe = \"tcp://localhost:" + subscribe + "\"",
      "",
      "[network]",
      "relax-checks = true",
      "",
      "[node]",
      "detect-drivers = true",
      "[server]",
      "registration-secret = \"feta\""
    };

    TomlConfig baseConfig = new TomlConfig(new StringReader(String.join("\n", rawConfig)));
    Config hubConfig = new CompoundConfig(
      new MapConfig(ImmutableMap.of("events", ImmutableMap.of("bind", true))),
      baseConfig);

    Server<?> hubServer = new Hub().asServer(setRandomPort(hubConfig)).start();

    waitUntilReady(hubServer, Boolean.FALSE);

    return hubServer;
  }

  private static Server<?> createStandalone() {
    String[] rawConfig = new String[]{
      "[network]",
      "relax-checks = true",
      "[node]",
      "detect-drivers = true",
      "[server]",
      "port = " + standalonePort,
      "registration-secret = \"provolone\""
    };
    Config config = new MemoizedConfig(
      new TomlConfig(new StringReader(String.join("\n", rawConfig))));

    Server<?> server = new Standalone().asServer(config).start();

    waitUntilReady(server, Boolean.TRUE);

    return server;
  }

  private static void waitUntilReady(Server<?> server, Boolean state) {
    HttpClient client = HttpClient.Factory.createDefault().createClient(server.getUrl());

    new FluentWait<>(client)
      .withTimeout(Duration.ofSeconds(5))
      .until(c -> {
        HttpResponse response = c.execute(new HttpRequest(GET, "/status"));
        Map<String, Object> status = Values.get(response, MAP_TYPE);
        return state.equals(status.get("ready"));
      });
  }

  private static Config setRandomPort(Config config) {
    return new MemoizedConfig(
      new CompoundConfig(
        new MapConfig(ImmutableMap.of("server", ImmutableMap.of("port", hubPort))),
        config));
  }
}
