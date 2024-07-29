public interface BatchAndSendFunction<T, U> {
    CompletableFuture<U> batchAndSend(List<IdentifiedRequest<T>> identifiedRequests, String destination);
}
