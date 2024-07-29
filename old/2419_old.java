         */
        public Builder maxAttempts(int maxAttempts) {
            this.maxAttempts = maxAttempts;
            return this;
        }

        /**
         * Set a function to modify the waiting interval after a failure.
         *
         * @param f Function to modify the interval after a failure
         * @return the RetryConfig.Builder
         */
        public Builder intervalFunction(IntervalFunction f) {
            this.intervalFunction = f;
            return this;
        }

        /**
         * Constructs the actual strategy based on the properties set previously.
         */
        public RetryStrategy build() {
            return new RetryStrategyImpl(maxAttempts, intervalFunction);
        }
    }

}
