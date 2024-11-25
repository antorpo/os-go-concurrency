package pipeline

func panicProof(
	goFunc func(),
	onPanic func(panic interface{}),
	deferred func()) {
	go func() {
		defer func() {
			defer deferred()

			if p := recover(); p != nil {
				onPanic(p)
			}
		}()

		goFunc()
	}()
}
