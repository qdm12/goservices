# Go services

`goservices` is a Go package to help manage services.

For now, it is notably used in [Gluetun](https://github.com/qdm12/gluetun) and [qdm12/dns](https://github.com/qdm12/dns/tree/v2.0.0-beta) to run multiple servers and 'loops' in the same program.

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

- the code is **fully test covered**
- Zero dependency (except for tests with [`golang/mock`](https://github.com/golang/mock) and [`stretchr/testify`](https://github.com/stretchr/testify)) - [![gographs](https://gographs.io/badge.svg)](https://gographs.io/repo/github.com/qdm12/goservices?cluster=false)
- the Go API should be stable until a v1.0.0 release
- the Go API will be guaranteed stable from the v1.0.0 release
- the code is **linted** with `golangci-lint` and a lot of linters
- There is a CI pipeline to test, lint, check mocks and check documentation on every commit.

## Sequence of services

To start and stop a sequence of services, you can use the [`Sequence` type](https://github.com/qdm12/goservices/blob/main/sequence.go#L10).
Note it itself implements the `Service` interface, so you can nest it with other service management types, like `Group`.

```go
 ctx := context.Background()

 settings := goservices.SequenceSettings{
  ServicesStart: []goservices.Service{serviceA, serviceB},
  ServicesStop:  []goservices.Service{serviceB, serviceA},
 }
 sequence, err := goservices.NewSequence(settings)
 if err != nil {
  return fmt.Errorf("creating services sequence: %w", err)
 }


 runError, err := sequence.Start(ctx)
 if err != nil {
  return fmt.Errorf("starting services sequence: %w", err)
 }

 select {
 case err = <-runError:
  return fmt.Errorf("services sequence crashed: %w", err)
 case <-ctx.Done():
  err = sequence.Stop()
  if err != nil {
   return fmt.Errorf("stopping services sequence: %w", err)
  }
  return nil
 }
```

[🏃 runnable example](examples/sequence/main.go)

## Group of services

To start and stop a group of services all in parallel, you can use the [`Group` type](https://github.com/qdm12/goservices/blob/main/group.go#L10).
Note it itself implements the `Service` interface, so you can nest it with other service management types, like `Sequence`.

A simplistic example would be:

```go
 ctx := context.Background()

 settings := goservices.GroupSettings{
  Services: []goservices.Service{serviceA, serviceB},
 }
 group, err := goservices.NewGroup(settings)
 if err != nil {
  return fmt.Errorf("creating services group: %w", err)
 }

 runError, err := group.Start(ctx)
 if err != nil {
  return fmt.Errorf("starting services group: %w", err)
 }

 select {
 case err = <-runError:
  return fmt.Errorf("services group crashed: %w", err)
 case <-ctx.Done():
  err = group.Stop()
  if err != nil {
   return fmt.Errorf("stopping services group: %w", err)
  }
  return nil
 }
```

[🏃 runnable example](examples/group/main.go)

## Auto-restart a service

To automatically restart a service when it crashes, you can use the [`Restarter` type](https://github.com/qdm12/goservices/blob/main/restarter.go#L10).
Note it itself implements the `Service` interface, so you can nest it with other service management types, like `Sequence`.

```go
 ctx := context.Background()

 settings := goservices.RestarterSettings{
  Service: serviceToRestart,
 }
 restarter, err := goservices.NewRestarter(settings)
 if err != nil {
  return fmt.Errorf("creating restarter: %w", err)
 }

 runError, startErr := restarter.Start(ctx)
 if startErr != nil {
  return fmt.Errorf("starting restarter: %w", startErr)
 }

 select {
 case err = <-runError:
  return fmt.Errorf("restarter crashed: %w", err)
 case <-ctx.Done():
  err = restarter.Stop()
  if err != nil {
   return fmt.Errorf("stopping restarter: %w", err)
  }
  return nil
 }
```

[🏃 runnable example](examples/restarter/main.go)

## Create a service

You can implement yourself the interface.
A good thread safe example to follow would be the [httpserver](httpserver) service implementation.

**HOWEVER** this is tedious to get right especially with the many race conditions possible (i.e. what if the service crashes at the same time as it is stopped?).

This is why this library provides a `RunWrapper` which creates a service from a `RunFunction`:

```go
type RunFunction func(ctx context.Context,
 ready chan<- struct{}, runError, stopError chan<- error)
```

Please see the [documentation of the `RunFunction`](https://github.com/qdm12/goservices/blob/main/runwrapper.go#L9-L53) to know the details on how to implement it correctly.

A concrete example is the previous implementation of the [`httpserver`](https://github.com/qdm12/goservices/blob/68f98ba0a1f7dc5a258fda3b2a88d16e79b9bd26/httpserver/server.go) service which was using this `RunWrapper`.

## Pre-built services

This library provides a few pre-built services:

- [`httpserver`](httpserver)

## Main branch dependency graph

![gographs](https://gographs.io/graph/github.com/qdm12/goservices.svg)
