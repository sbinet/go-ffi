go-ffi
======

[![Build Status](https://secure.travis-ci.org/sbinet/go-ffi.png)](http://travis-ci.org/sbinet/go-ffi)

The ``ffi`` package wraps the ``libffi`` ``C`` library (and ``dlopen/dlclose``) to provide an easy way to call arbitrary functions from ``Go``.

Installation
------------

``ffi`` is go-get-able:

```
$ go get github.com/sbinet/go-ffi/pkg/ffi
```

Example
-------

``` go
// dl-open a library: here, the math library on unix
lib, err := ffi.NewLibrary("m")
handle_err(err)

// get a handle to 'cos', with the correct signature
cos, err := lib.Fct("cos", ffi.Double, []Type{ffi.Double})
handle_err(err)

// call it
out := cos(0.).Float()
println("cos(0.)=", out)

err = lib.Close()
handle_err(err)
```

Limitations/TODO
-----------------

- no check is performed b/w what the user provides as a signature and the "real" signature

- it would be handy to use some tool to automatically infer the "real" function signature

- it would be handy to also handle structs

- better handling of types with no direct equivalent in go
  (short,void,...)

- better handling of C_string and conversion to/from Go strings

Documentation
-------------

http://go.pkgdoc.org/github.com/sbinet/go-ffi/pkg/ffi

