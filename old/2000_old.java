        RetryStrategies.linear(new RetryStrategies.Retryable() {
            @Override
            public void call() {
                AirplaneMode.disable();
            }
        }, new RetryStrategies.Condition() {
            @Override
            public boolean isMet() {
                return InternetConnectivity.isOnline();
            }
        }, 3, 2);
    }

    public static void goOffline() {
        RetryStrategies.linear(new RetryStrategies.Retryable() {
            @Override
            public void call() {
                AirplaneMode.enable();
            }
        }, new RetryStrategies.Condition() {
            @Override
            public boolean isMet() {
                return InternetConnectivity.isOffline();
            }
        }, 3, 2);
    }

    private static boolean isOnline() {
        int connectionTimeoutMs = (int) TimeUnit.SECONDS.toMillis(2);
        String host = "amazon.com";
        int port = 443;

        Socket socket = new Socket();
        try {
            socket.connect(new InetSocketAddress(host, port), connectionTimeoutMs);
            socket.close();
            return true;
        } catch (IOException socketFailure) {
            return false;
        }
    }

    private static boolean isOffline() {
        return !isOnline();
    }
}