package helpers

import "fmt"

func NewPrintHooks() *PrintHooks {
	return &PrintHooks{}
}

type PrintHooks struct{}

func (h *PrintHooks) OnStart(service string) { fmt.Println("Starting", service) }
func (h *PrintHooks) OnStarted(service string, err error) {
	fmt.Println("Started", service, "with error", err)
}
func (h *PrintHooks) OnStop(service string) { fmt.Println("Stopping", service) }
func (h *PrintHooks) OnStopped(service string, err error) {
	fmt.Println("Stopped", service, "with error", err)
}
func (h *PrintHooks) OnCrash(service string, err error) {
	fmt.Println("Crashed", service, "with error", err)
}
