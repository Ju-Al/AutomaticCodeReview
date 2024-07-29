	val := ctx.Value(localStorageKey)

	if val == nil {
		panic("This isn't the context for a pipeline spawned go routine, or the LocalStorage was deleted")
	}

	if ls, ok := val.(LocalStorage); !ok {
		panic("This isn't the context for a pipeline spawned go routine, or the LocalStorage was deleted")
	} else {
		return ls
	}
}
