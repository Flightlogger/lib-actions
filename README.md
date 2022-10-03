# lib-actions
Golang library for FlightLogger GitHub actions.

## Publishing

To publish a new version of the library to pkg.go.dev, simply run the `publish.sh` script with the version number as the first and only argument:
```
$ ./publish.sh v4.2.0
```

_NOTE: Golang requires module versions to follow the exact format: `vX.X.X`, where `X` is an integer._
