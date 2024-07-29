    {
        private readonly IFlushAwaiter _awaiter;

        public FlushAsyncAwaitable(IFlushAwaiter awaiter)
        {
            _awaiter = awaiter;
        }

        public bool IsCompleted => _awaiter.IsCompleted;

        public bool GetResult() => _awaiter.GetResult();

        public FlushAsyncAwaitable GetAwaiter() => this;

        public void UnsafeOnCompleted(Action continuation) => _awaiter.OnCompleted(continuation);

        public void OnCompleted(Action continuation) => _awaiter.OnCompleted(continuation);
    }
}
