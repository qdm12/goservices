// Example of a sequence of services with goservices.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/qdm12/goservices"
	"github.com/qdm12/goservices/examples/helpers"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := runServices(ctx)
	stop()
	if err != nil {
		log.Fatal(err)
	}
}

func runServices(ctx context.Context) (err error) {
	serviceA := helpers.NewDummyService(
		helpers.DummyServiceSettings{Name: "A"},
	)
	serviceB := helpers.NewDummyService(
		helpers.DummyServiceSettings{Name: "B"},
	)
	settings := goservices.SequenceSettings{
		ServicesStart: []goservices.Service{serviceA, serviceB},
		ServicesStop:  []goservices.Service{serviceB, serviceA},
		Hooks:         helpers.NewPrintHooks(),
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
}
