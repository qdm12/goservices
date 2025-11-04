// Package helpers provides helper implementations for examples.
package helpers

import "fmt"

// NewPrintHooks creates a new PrintHooks instance.
func NewPrintHooks() *PrintHooks {
	return &PrintHooks{}
}

// PrintHooks implements service handler hooks.
type PrintHooks struct{}

// OnStart prints the service starting.
func (h *PrintHooks) OnStart(service string) { fmt.Println("Starting", service) }

// OnStarted prints the service started, with its eventual error.
func (h *PrintHooks) OnStarted(service string, err error) {
	fmt.Println("Started", service, "with error", err)
}

// OnStop prints the service stopping.
func (h *PrintHooks) OnStop(service string) { fmt.Println("Stopping", service) }

// OnStopped prints the service stopped, with its eventual error.
func (h *PrintHooks) OnStopped(service string, err error) {
	fmt.Println("Stopped", service, "with error", err)
}

// OnCrash prints the service crashing with its error.
func (h *PrintHooks) OnCrash(service string, err error) {
	fmt.Println("Crashed", service, "with error", err)
}
