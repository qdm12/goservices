# Design

## Service interface

The first thing to determine was what interactions are needed between a parent caller and a service it manages?
These are the following:

- Know when the service is ready to be used
- Know when and why the service crashes
- Signal the service to stop
- Know when the service is stopped and if it stopped with an error

Given these points, two main designs emerged for a `Service` interface:

1. A single `Run` blocking function:

    ```go
    Run(ctx context.Context, ready chan<- struct{}, runError, stopError chan<- error)
    ```

    It would:
    - close the `ready` channel once the service is ready
    - write to the `runError` channel when an unexpected fatal error occurs
    - exit on `ctx` cancellation and write an eventual error to the `stopError` channel.

2. Two functions, one to start the service and one to stop it:

    ```go
    Start(ctx context.Context) (runError <-chan error, err error)
    Stop() error
    ```

    - `Start` starts the service:
      - Its `ctx` argument is used by the caller to cancel the start operation only
      - It returns `runError`, a ready only error channel the caller must listen on to detect when and why the service crashed
      - It returns `err`, an error which is a **start error** only
    - `Stop` stops the service, and returns an error in case stopping the service failed.

    Note `Stop` does not take a context argument since this would allow the service to be left in a bad state (half-way stopped) which makes service implementation a nightmare.

Between the two designs, we can represent the following equivalence table:

| Service interface | Run | Start + Stop |
| --- | --- | --- |
| Readiness | read only channel `ready` injected to `Run` | `Start` returns a nil `err` |
| Crash detection | read only channel `runError` injected to `Run` | read only channel `runError` returned by `Start` |
| Stop signal | `ctx` argument injected to `Run` | `ctx` argument injected to `Start` and then calling `Stop` |
| Stopped detection | ready only channel `stopError` injected to `Run` | `Stop` returns an error |

The second design *Start + Stop* has a some advantages compared to the first design *Run*:

- it distinguishes between a start error (`err` returned by `Start`) and a run error (error written ton `runError`)
- the caller cannot write to channels since these are always returned as read-only
- the caller does not need to launch the service in a goroutine, this is abstracted away

On the other hand, one major disadvantage is that thread safety is much harder to achieve since it's two non-blocking functions `Start` and `Stop`, against one blocking `Run` function.

The second design *Start + Stop* was chosen due to **its ease of user for a caller**, eventhough it is harder to implement correctly if thread safety matters.

The last part was to add a `String() string` function to the interface to allow for a service to be identified by a string. This is especially useful for error messages and logs.

The final `Service` interface is the following:

```go
Start(ctx context.Context) (runError <-chan error, err error)
Stop() error
String() string
```

## Service implementation

[👉 http server service example](httpserver)

The `Service` interface can be implemented rather simply, but it gets more complicated when it has to be thread safe, which it should be ideally.

The following discusses how to implement a service in a thread safe manner.

The basic service fields you would need are:

```go
 startStopMutex        sync.Mutex
 state                 goservices.State
 stateMutex            sync.RWMutex
 runCtx                context.Context
 runCancel             context.CancelFunc
 runDone               <-chan error
```

- `startStopMutex` is used to prevent the service from being started and stopped at the same time
- `state` is the current state of the service, and we use the exported `goservices.State` type for this, which can be one of:
  - `goservices.StateStopped` (default zero value)
  - `goservices.StateStopping`
  - `goservices.StateStarting`
  - `goservices.StateStarted`
  - `goservices.StateCrashed`
- `stateMutex` is used to protect the state field from data races
- `runCtx` and `runCancel` are used to cancel the underlying running loop of the service
- `runDone` is used to signal when the running loop has exited and if it failed to exit.

The `Start` function can be implemented as follows:

```go
func (s *Service) Start(ctx context.Context) (runError <-chan error, err error) {
 // Lock the startStopMutex to prevent concurrent calls
 // to `Start` and `Stop`.
 s.startStopMutex.Lock()
 defer s.startStopMutex.Unlock()

 // Lock the stateMutex in case the service is already running
 // and tries to change the state to StateCrashed.
 s.stateMutex.RLock()
 state := s.state
 // no need to keep a lock on the state since the `startStopMutex`
 // prevents concurrent calls to `Start` and `Stop`.
 s.stateMutex.RUnlock()
 if state == goservices.StateRunning {
  // service is already running, the caller should not call
  // Start twice on the same service, so an error is returned.
  // This one should always wrap `goservices.ErrAlreadyStarted`
  // so the caller can ignore it using `errors.Is()` if needed.
  return nil, fmt.Errorf("%s: %w", s, goservices.ErrAlreadyStarted)
 }

 s.state = goservices.StateStarting

 err := doSomeSetup(ctx)
 if err != nil {
  s.state = goservices.StateCrashed
  return nil, err
 }

 // Hold the state mutex locked in case the running goroutine
 // crashes instantly at start.
 s.stateMutex.Lock()

 s.runCtx, s.runCancel = context.WithCancel(context.Background())
 runErrorBiDirectional := make(chan error)
 runError = runErrorBiDirectional
 ready := make(chan struct{})
 runDoneBiDirectional := make(chan error)
 s.runDone = runDoneBiDirectional
 go func() {
  // it takes time to launch a goroutine, so
  // use a ready channel to signal the parent
  // goroutine it is ready to be used.
  close(ready)

  // Run an infinite loop until the runCtx is canceled.
  for s.runCtx.Err() != nil {
    err := doSomeSynchronousWork()
    if err != nil {
      s.stateMutex.Lock()
      s.state = goservices.StateCrashed
      s.stateMutex.Unlock()
      _ = doSomeCleanup()
      runErrorBiDirectional <- err
      return // exit the goroutine
    }
  }

  // Do some cleanup before exiting the run loop
  // This can only be triggered by a call to `Stop`.
  err := doSomeCleanup()
  runDoneBiDirectional <- err
 }()

 <-ready
 s.state = goservices.StateRunning
 s.stateMutex.Unlock()

 return runError, nil
}
```

and the `Stop` function:

```go
func (s *Service) Stop() (err error) {
 // Lock the startStopMutex to prevent concurrent calls
 // to `Start` and `Stop`.
 s.startStopMutex.Lock()
 defer s.startStopMutex.Unlock()

 // Lock the stateMutex in case the service is already running
 // and tries to change the state to StateCrashed.
 s.stateMutex.Lock()
 switch s.state {
 case goservices.StateRunning:
  // continue stopping the sequence
 case goservices.StateCrashed:
   // service is already stopped
  s.stateMutex.Unlock()
  return nil
 case goservices.StateStopped:
  // service is already stopped, the caller should not call
  // Stop twice on the same service, so an error is returned.
  // This one should always wrap `goservices.ErrAlreadyStopped`
  // so the caller can ignore it using `errors.Is()` if needed.
  s.stateMutex.Unlock()
  return fmt.Errorf("%s: %w", s, goservices.ErrAlreadyStopped)
 case goservices.StateStarting, goservices.StateStopping:
  panic("unreachable")
 }

 s.state = goservices.StateStopping

 // Unlock the stateMutex, so the state from this point
 // can be StateCrashed for a very short time, before
 // becoming StateStopped.
 s.stateMutex.Unlock()

 s.runCancel()
 err = <-s.runDone

 s.state = goservices.StateStopped
 return err
}
```
