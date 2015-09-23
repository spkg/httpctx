package log_test

import (
	"golang.org/x/net/context"
	"sp.com.au/exp/log"
)

type Parameters []log.Parameter

func messingAround(ctx context.Context, params Parameters) {

}

func moreMessingAround() {
	ctx := context.Background()
	messingAround(ctx, Parameters{{"Name", 1}, {"P2", "xxx"}})
	messingAround(ctx, Parameters{{"Name", 1}})
}

func ExampleOption(ctx context.Context, n1, n2 int) error {
	if err := doSomethingWith(n1, n2); err != nil {
		return log.Error("cannot doSomething",
			log.WithValue("n1", n1),
			log.WithValue("n2", n2),
			log.WithError(err),
			log.WithContext(ctx))
	}

	// .. more processing and then ...

	return nil
}

func ExampleWithValue(n1, n2 int) error {
	if err := doSomethingWith(n1, n2); err != nil {
		return log.Error("doSomethingWith failed",
			log.WithValue("n1", n1),
			log.WithValue("n2", n2))
	}

	// ... more processing and then ...

	return nil
}

func ExampleWithError() error {
	if err := doSomething(); err != nil {
		return log.Error("doSomething failed",
			log.WithError(err))
	}

	// ... more processing and then ...

	return nil
}

func ExampleNewContext(ctx context.Context, n1, n2 int) error {
	ctx = log.NewContext(ctx, "n1", n1)
	ctx = log.NewContext(ctx, "n2", n2)

	if err := doSomethingWith(n1, n2); err != nil {
		return log.Error("doSomethingWith failed",
			log.WithError(err),
			log.WithContext(ctx))
	}

	log.Debug("did something with",
		log.WithContext(ctx))

	// ... more processing and then ...

	return nil
}

func doSomethingWith(n1 int, n2 int) error {
	return nil
}

func doSomething() error {
	return nil
}

func doAnotherThing() error {
	return nil
}

func doOneMoreThing() error {
	return nil
}
