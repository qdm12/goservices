# Go services

`goservices` is a Go package to help manage services.

🚧 Logo to be added 🚧

What is a service? It's this interface currently:

```go
type Service interface {
 // String returns the service name.
 // It is assumed to be constant over the lifetime of the service.
 String() string
 // Start starts the service.
 // On success, it returns a run error channel and a nil error.
 // On failure, it returns a nil run error channel and an error.
 // If the service crashes, only one single error should be sent in
 // the error channel.
 // When the service is stopped, the service should NOT send an error
 // in the run error channel or close this one.
 // Start takes in a context and the implementation should promptly return
 // the context error wrapped in `startErr` if the context is canceled.
 Start(ctx context.Context) (runError <-chan error, startErr error)
 // Stops stops the service.
 // A service should NOT close or write an error to its run error channel
 // if it is stopped.
 Stop() (err error)
}
```

## Stability

- the code is **fully test covered** (99.6% - 0.4% being unreachable panic cases)
- Zero dependency (except for tests)
- the Go API is NOT guaranteed to be stable yet, but it should stay stable for a while
- the code is **linted** with `golangci-lint` and a lot of linters

## Sequence of services

To start and stop a sequence of services, you can use the [`Sequence` type](https://github.com/qdm12/goservices/blob/468bda9ee482fcaca953b1b63b6cdabf8b1aa6a6/sequence.go#L10).
Note it itself implements the `Service` interface, so you can nest it with other service management types, like `Group`.

🚧 Examples to be added 🚧

## Group of services

To start and stop a group of services all in parallel, you can use the [`Group` type](https://github.com/qdm12/goservices/blob/468bda9ee482fcaca953b1b63b6cdabf8b1aa6a6/group.go#L10).
Note it itself implements the `Service` interface, so you can nest it with other service management types, like `Sequence`.

🚧 Examples to be added 🚧

## Auto-restart a service

To automatically restart a service when it crashes, you can use the [`Restarter` type](https://github.com/qdm12/goservices/blob/468bda9ee482fcaca953b1b63b6cdabf8b1aa6a6/restarter.go#L10).
Note it itself implements the `Service` interface, so you can nest it with other service management types, like `Sequence`.

🚧 Examples to be added 🚧

## Create a service

You can implement yourself the interface.

**HOWEVER** this is tedious to get right especially with the many race conditions possible (i.e. what if the service crashes at the same time as it is stopped?).

This is why this library provides a `RunWrapper` which creates a service from a `RunFunction`:

```go
type RunFunction func(ctx context.Context,
 ready chan<- struct{}, runError, stopError chan<- error)
```

Please see the [documentation of the `RunFunction`](https://github.com/qdm12/goservices/blob/468bda9ee482fcaca953b1b63b6cdabf8b1aa6a6/runwrapper.go#L9-L53) to know the details on how to implement it correctly.

## Pre-built services

This library provides a few pre-built services:

- [`httpserver`](httpserver)
