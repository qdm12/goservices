package goservices

import (
	"fmt"

	"github.com/qdm12/goservices/hooks"
)

// RestarterSettings contains settings for a restarter.
type RestarterSettings struct {
	// Service is the service to restart.
	// It must be set for settings validation to succeed.
	Service Service
	// Hooks are hooks to call when the service starts,
	// stops or crashes. It defaults to a noop hooks
	// implementation.
	Hooks Hooks
}

// SetDefaults sets the defaults for the restarter settings.
func (r *RestarterSettings) SetDefaults() {
	if r.Hooks == nil {
		r.Hooks = hooks.NewNoop()
	}
}

// Validate validates the restarter settings.
func (r RestarterSettings) Validate() (err error) {
	if r.Service == nil {
		return fmt.Errorf("%w", ErrNoService)
	}

	return nil
}
