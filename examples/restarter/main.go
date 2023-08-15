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
	err := runRestarter(ctx)
	stop()
	if err != nil {
		log.Fatal(err)
	}
}

func runRestarter(ctx context.Context) (err error) {
	settings := goservices.RestarterSettings{
		Service: helpers.NewDummyService(
			helpers.DummyServiceSettings{Name: "A", MaxLife: time.Second},
		),
		Hooks: helpers.NewPrintHooks(),
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
}
