# Context-Aware HTTP Handlers

[![GoDoc](https://godoc.org/github.com/spkg/httpctx?status.svg)](https://godoc.org/github.com/spkg/httpctx)
[![Build Status](https://travis-ci.org/spkg/httpctx.svg?branch=master)](https://travis-ci.org/spkg/httpctx)
[![license](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/spkg/httpctx/master/license.txt)
[![Coverage](http://gocover.io/_badge/github.com/spkg/httpctx)](http://gocover.io/github.com/spkg/httpctx)




The `httpctx` package provides a convenient way to handle HTTP requests
using "context-aware" handler functions. Context-aware handlers
are different from the standard `http.Handler` in two important ways:

1. They accept an additional parameter of the (almost) standard type `context.Context`
(golang.org/x/net/context); and
2. They return an error result.

The `httpctx` package also implements a simple middleware chaining mechanism. The idea
for this comes from Justinas Stankevičius and his `alice` package. (https://github.com/justinas/alice).

For usage examples, refer to the [GoDoc](https://godoc.org/github.com/spkg/httpctx) documentation.
