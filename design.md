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
