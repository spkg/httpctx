# Context-Aware HTTP Handlers

[![GoDoc](https://godoc.org/github.com/spkg/httpctx?status.svg)](https://godoc.org/github.com/spkg/httpctx)

The `httpctx` package provides a convenient way to handle HTTP requests
using "context-aware" handler functions. Context-aware handlers
are different from the standard `http.Handler` in two important ways:

1. They accept an additional parameter of the (almost) standard type `context.Context`
(golang.org/x/net/context); and
2. They return an error result.

The `httpctx` package also implements a simple middleware chaining mechanism. The idea
for this comes from Justinas Stankevičius and his `alice` package. (https://github.com/justinas/alice).

For usage examples, refer to the [GoDoc](https://godoc.org/github.com/spkg/httpctx) documentation.
