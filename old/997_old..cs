        {
            await ExecuteAsync(null).ConfigureAwait(false);
        }

        public async Task ExecuteAsync(object parameter)
        {
            if (DoCanExecute(parameter))
            {
                await ExecuteConcurrentAsync(parameter).ConfigureAwait(false);
            }
        }

        private async Task ExecuteConcurrentAsync(object parameter)
        {
            bool started = false;
            try
            {
                lock (_syncRoot)
                {
                    if (_concurrentExecutions == 0)
                    {
                        InitCancellationTokenSource();
                    }
                    else if (!_allowConcurrentExecutions)
                    {
                        Mvx.Trace(MvxTraceLevel.Diagnostic, "MvxAsyncCommand : execute ignored, already running.");
                        return;
                    }
                    _concurrentExecutions++;
                    started = true;
                }
                if (!_allowConcurrentExecutions)
                {
                    RaiseCanExecuteChanged();
                }
                // With configure await false, the CanExecuteChanged raised in finally clause might run in another thread.
                // This should not be an issue as long as ShouldAlwaysRaiseCECOnUserInterfaceThread is true.
                await ExecuteWithErrorHandlingAsync(parameter).ConfigureAwait(false);
            }
            finally
            {
                if (started)
                {
                    lock (_syncRoot)
                    {
                        _concurrentExecutions--;
                        if (_concurrentExecutions == 0)
                        {
                            ClearCancellationTokenSource();
                        }
                    }
                    if (!_allowConcurrentExecutions)
                    {
                        RaiseCanExecuteChanged();
                    }
                }
            }
        }
        
        private async Task ExecuteWithErrorHandlingAsync(object parameter)
        {
            try
            {
                if (!CancelToken.IsCancellationRequested)
                {
                    // ConfigureAwait(false) => Same issue as RaiseCanExecuteChanged
                    // Not a problem when ShouldAlwaysRaiseCECOnUserInterfaceThread is true
                    await DoExecuteAsync(parameter).ConfigureAwait(false);
                }
            }
            // Uncomment to avoid reporting cancellation as error
            //catch(TaskCanceledException)
            //{ }
            catch (Exception e)
            {
                Mvx.Trace(MvxTraceLevel.Error, "MvxAsyncCommand : exception executing task : ", e);
                OnErrorOccured(new MvxCommandErrorEventArgs(e));
            }
        }

        protected virtual void OnErrorOccured(MvxCommandErrorEventArgs e)
        {
            var tmp = ErrorOccured;
            if (tmp != null)
            {
                if (ShouldAlwaysRaiseCECOnUserInterfaceThread)
                {
                    InvokeOnMainThread(() => tmp(this, e));
                }
                else
                {
                    tmp(this, e);
                }
            }
        }

        private void ClearCancellationTokenSource()
        {
            if (_cts == null)
            {
                Mvx.Error("MvxAsyncCommand : Unexpected ClearCancellationTokenSource, no token available!");
            }
            else
            {
                _cts.Dispose();
                _cts = null;
            }
        }

        private void InitCancellationTokenSource()
        {
            if (_cts != null)
            {
                Mvx.Error("MvxAsyncCommand : Unexpected InitCancellationTokenSource, a token is already available!");
            }
            _cts = new CancellationTokenSource();
        }
    }

    public class MvxAsyncCommand
        : MvxAsyncCommandBase
        , IMvxCommand
    {
        private readonly Func<CancellationToken, Task> _execute;
        private readonly Func<bool> _canExecute;

        public MvxAsyncCommand(Func<Task> execute, Func<bool> canExecute = null, bool allowConcurrentExecutions = false)
            : base(allowConcurrentExecutions)
        {
            if (execute == null)
                throw new ArgumentNullException("execute");

            _execute = (cancellationToken) => execute();
            _canExecute = canExecute;
        }

        public MvxAsyncCommand(Func<CancellationToken, Task> execute, Func<bool> canExecute = null, bool allowConcurrentExecutions = false)
            : base(allowConcurrentExecutions)
        {
            if (execute == null)
                throw new ArgumentNullException("execute");

            _execute = execute;
            _canExecute = canExecute;
        }

        protected override bool DoCanExecute(object parameter)
        {
            return _canExecute == null || _canExecute();
        }

        protected override Task DoExecuteAsync(object parameter)
        {
            return _execute(CancelToken);
        }

        public static MvxAsyncCommand<T> CreateCommand<T>(Func<T, Task> execute, Func<T, bool> canExecute = null, bool allowConcurrentExecutions = false)
        {
            return new MvxAsyncCommand<T>(execute, canExecute, allowConcurrentExecutions);
        }

        public static MvxAsyncCommand<T> CreateCommand<T>(Func<T, CancellationToken, Task> execute, Func<T, bool> canExecute = null, bool allowConcurrentExecutions = false)
        {
            return new MvxAsyncCommand<T>(execute, canExecute, allowConcurrentExecutions);
        }
    }

    public class MvxAsyncCommand<T>
        : MvxAsyncCommandBase
        , IMvxCommand
    {
        private readonly Func<T, CancellationToken, Task> _execute;
        private readonly Func<T, bool> _canExecute;

        public MvxAsyncCommand(Func<T, Task> execute, Func<T, bool> canExecute = null, bool allowConcurrentExecutions = false)
            : base(allowConcurrentExecutions)
        {
            if (execute == null)
                throw new ArgumentNullException("execute");

            _execute = (p, c) => execute(p);
            _canExecute = canExecute;
        }

        public MvxAsyncCommand(Func<T, CancellationToken, Task> execute, Func<T, bool> canExecute = null, bool allowConcurrentExecutions = false)
            : base(allowConcurrentExecutions)
        {
            if (execute == null)
                throw new ArgumentNullException("execute");

            _execute = execute;
            _canExecute = canExecute;
        }

        protected override bool DoCanExecute(object parameter)
        {
            return _canExecute == null || _canExecute((T)typeof(T).MakeSafeValueCore(parameter));
        }

        protected override Task DoExecuteAsync(object parameter)
        {
            return _execute((T)typeof(T).MakeSafeValueCore(parameter), CancelToken);
        }
    }
}