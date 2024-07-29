        throw new IllegalStateException("Unable to establish connection to ActiveMQ server", e);
      }
      return connection;
    }

    void close() throws IOException {
      Connection maybeConnection = connection;
      if (maybeConnection != null) {
        try {
          maybeConnection.close();
        }catch (Exception e){
          throw new IOException(e);
        }

      }
    }

    Connection get() {
      if (connection == null) {
        synchronized (this) {
          if (connection == null) {
            connection = compute();
          }
        }
      }
      return connection;
    }


  }


  /**
   * Consumes spans from messages on a ActiveMQ queue. Malformed messages will be discarded. Errors
   * in the storage component will similarly be ignored, with no retry of the message.
   */
  static class ActiveMQSpanConsumerMessageListener implements MessageListener {
    final Collector collector;
    final CollectorMetrics metrics;

    ActiveMQSpanConsumerMessageListener(Collector collector, CollectorMetrics metrics) {
      this.collector = collector;
      this.metrics = metrics;
    }

    @Override
    public void onMessage(Message message) {
      metrics.incrementMessages();
      try {
        if(message instanceof BytesMessage) {
            byte [] data = new byte[(int)((BytesMessage)message).getBodyLength()];
            ((BytesMessage)message).readBytes(data);
            this.collector.acceptSpans(data, NOOP);
        }else if(message instanceof TextMessage){
          String text = ((TextMessage)message).getText();
          byte [] data = text.getBytes();
          this.collector.acceptSpans(data, NOOP);
        }
      }catch (Exception e){
      }
    }


  }

}
