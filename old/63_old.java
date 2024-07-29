      LOG.info("SentryAppenderBundle is installed, but there is no sentry configuration ");
      return;
    }

    if(!maybeSentryConfiguration.get().getDsn().isPresent()){
      LOG.warn("SentryAppenderBundle is installed and sentry config exists but DSN is missing");
      return;
    }
    
    String dsn = maybeSentryConfiguration.get().getDsn().get();
    Raven raven = RavenFactory.ravenInstance(dsn);
    SentryAppender sentryAppender = new SentryAppender(raven);

    Logger rootLogger = (Logger) LoggerFactory.getLogger(Logger.ROOT_LOGGER_NAME);

    sentryAppender.start();

    rootLogger.addAppender(sentryAppender);
  }

  public void initialize(Bootstrap<?> bootstrap) {
  }
}
