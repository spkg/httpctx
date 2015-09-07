package log_test

import (
	"sp.com.au/exp/log"
)

func ExampleErr() error {
	if err := doSomething(); err != nil {
		// log an error
		return log.Err(err)
	}

	if err := doAnotherThing(); err != nil {
		// this is a warning, not an error
		return log.Err(err,
			log.WithSeverity(log.SeverityWarning))
	}

	if err := doOneMoreThing(); err != nil {
		return log.Error("doOneMoreThingFailed",
			log.WithError(err))
	}

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
