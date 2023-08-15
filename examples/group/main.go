package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

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
	settings := goservices.GroupSettings{
		Services: []goservices.Service{
			helpers.NewDummyService(
				helpers.DummyServiceSettings{Name: "A", MaxStart: time.Second},
			),
			helpers.NewDummyService(
				helpers.DummyServiceSettings{Name: "B", MaxStart: time.Second},
			),
		},
		Hooks: helpers.NewPrintHooks(),
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
}
