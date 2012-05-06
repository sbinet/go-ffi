go-ffi
======

The ``ffi`` package wraps the ``libffi`` ``C`` library (and ``dlopen/dlclose``) to provide an easy way to call arbitrary functions from ``Go``.

Installation
------------

``ffi`` is go-get-able::

  $ go get bitbucket.org/binet/go-ffi/pkg/ffi


Example
-------

::

  // dl-open a library: here, libm on macosx
  lib, err := ffi.NewLibrary("libm.dylib")
  handle_err(err)

  // get a handle to 'cos', with the correct signature
  cos, err := lib.Fct("cos", ffi.Double, []Type{ffi.Double})
  handle_err(err)

  // call it
  out := cos(0.).Float()
  println("cos(0.)=", out)

  err = lib.Close()
  handle_err(err)

Limitations/TODO
-----------------

- no check is performed b/w what the user provides as a signature and the "real" signature

- it would be handy to just provide the name of the library (ie: "m") instead of its filename (ie: "libm.dylib")

- it would be handy to use some tool to automatically infer the "real" function signature

- it would be handy to also handle structs


Documentation
-------------

http://go.pkgdoc.org/bitbucket.org/binet/go-ffi/pkg/ffi

