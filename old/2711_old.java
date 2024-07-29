          CAPTURED_REQUESTS.add(agg);
          responseFuture.complete(HttpResponse.of(YELLOW_RESPONSE));
        }).exceptionally(t -> {
          responseFuture.completeExceptionally(t);
          return null;
        });
        return HttpResponse.from(responseFuture);
      });
    }
  };

  @Configuration static class TlsSelfSignedConfiguration {
    @Bean @Qualifier(QUALIFIER)
    ClientFactory zipkinElasticsearchClientFactory() {
      return new ClientFactoryBuilder().sslContextCustomizer(
        ssl -> ssl.trustManager(InsecureTrustManagerFactory.INSTANCE)).build();
    }
  }

  AnnotationConfigApplicationContext context = new AnnotationConfigApplicationContext();
  ElasticsearchStorage storage;

  @Before public void init() {
    TestPropertyValues.of(
      "spring.config.name=zipkin-server",
      "zipkin.storage.type:elasticsearch",
      "zipkin.storage.elasticsearch.username:Aladdin",
      "zipkin.storage.elasticsearch.password:OpenSesame",
      // force health check usage
      "zipkin.storage.elasticsearch.hosts:https://127.0.0.1:1234,https://127.0.0.1:"
        + server.httpsPort())
      .applyTo(context);
    context.register(
      TlsSelfSignedConfiguration.class,
      PropertyPlaceholderAutoConfiguration.class,
      ZipkinElasticsearchStorageConfiguration.class);
    context.refresh();
    storage = context.getBean(ElasticsearchStorage.class);
  }

  @After public void close() {
    CAPTURED_REQUESTS.clear();
  }

  @Test public void healthcheck_usesAuthAndTls() throws Exception {
    assertThat(storage.check().ok()).isTrue();

    // This loops to avoid a race condition
    boolean indexHealth = false, clusterHealth = false;
    while (!indexHealth || !clusterHealth) {
      AggregatedHttpRequest next = CAPTURED_REQUESTS.take();
      // hard coded for sanity taken from https://en.wikipedia.org/wiki/Basic_access_authentication
      assertThat(next.headers().get("Authorization"))
        .isEqualTo("Basic QWxhZGRpbjpPcGVuU2VzYW1l");

      String path = next.path();
      if (path.equals("/_cluster/health")) {
        clusterHealth = true;
      } else if (path.equals("/_cluster/health/zipkin*span-*")) {
        indexHealth = true;
      } else {
        fail("unexpected request to " + path);
      }
    }
  }
}
