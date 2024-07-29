        if (message == null) {
            return false;
        }

        T item = projectionFn.apply(message);
        if (item == null) {
            return false;
        }

        if (tryEmit(item)) {
            flushFn.accept(session);
        } else {
            pendingItem = item;
        }
        return false;
    }

    /**
     * @return {@code true} if emitting pending item is successful
     * {@code false} otherwise
     */
    private boolean emitPendingItem() {
        if (pendingItem == null) {
            return true;
        }
        if (tryEmit(pendingItem)) {
            pendingItem = null;
            flushFn.accept(session);
            return true;
        }
        return false;
    }

    @Override
    public void close(@Nullable Throwable error) {
        closeIfNotNull(consumer);
        closeIfNotNull(session);
    }

    private static void closeIfNotNull(AutoCloseable closeable) {
        if (closeable != null) {
            uncheckRun(closeable::close);
        }
    }

    private static final class Supplier<T> implements ProcessorSupplier {

        static final long serialVersionUID = 1L;

        private final DistributedSupplier<Connection> connectionSupplier;
        private final DistributedFunction<Connection, Session> sessionFn;
        private final DistributedFunction<Session, MessageConsumer> consumerFn;
        private final DistributedConsumer<Session> flushFn;
        private final DistributedFunction<Message, T> projectionFn;

        private transient Connection connection;

        private Supplier(DistributedSupplier<Connection> connectionSupplier,
                         DistributedFunction<Connection, Session> sessionFn,
                         DistributedFunction<Session, MessageConsumer> consumerFn,
                         DistributedConsumer<Session> flushFn,
                         DistributedFunction<Message, T> projectionFn) {
            this.connectionSupplier = connectionSupplier;
            this.sessionFn = sessionFn;
            this.consumerFn = consumerFn;
            this.flushFn = flushFn;
            this.projectionFn = projectionFn;
        }

        @Override
        public void init(@Nonnull Context context) {
            connection = connectionSupplier.get();
            uncheckRun(() -> connection.start());
        }

        @Override
        public void close(@Nullable Throwable error) {
            closeIfNotNull(connection);
        }

        @Nonnull
        @Override
        public Collection<? extends Processor> get(int count) {
            return range(0, count)
                    .mapToObj(i -> new StreamJmsP<>(connection, sessionFn, consumerFn, flushFn, projectionFn))
                    .collect(Collectors.toList());
        }
    }
}
