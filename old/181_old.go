	assert.True(form, "error message has expected prefix for timeouts, got %q", err.Error())
	_, ok := err.(transport.TimeoutError)
	assert.True(ok, "transport should be a TimeoutError")
}

func newTestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	return ctx
}
