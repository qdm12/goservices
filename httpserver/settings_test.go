package httpserver

import (
	"net/http"
	reflect "reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Settings_SetDefaults(t *testing.T) {
	t.Parallel()

	cancelHandler := func() {}

	testCases := map[string]struct {
		settings         Settings
		expectedSettings Settings
	}{
		"empty settings": {
			expectedSettings: Settings{
				Name:              stringPtr(""),
				Address:           stringPtr(""),
				ShutdownTimeout:   3 * time.Second,
				ReadTimeout:       10 * time.Second,
				ReadHeaderTimeout: time.Second,
				Logger:            &noopLogger{},
				CancelHandler:     func() {},
			},
		},
		"all settings fields set": {
			settings: Settings{
				Name:              stringPtr("x"),
				Handler:           http.NewServeMux(),
				Address:           stringPtr("test"),
				ReadTimeout:       time.Second,
				ReadHeaderTimeout: 2 * time.Second,
				ShutdownTimeout:   3 * time.Second,
				Logger:            NewMockInfoer(nil),
				CancelHandler:     cancelHandler,
			},
			expectedSettings: Settings{
				Name:              stringPtr("x"),
				Handler:           http.NewServeMux(),
				Address:           stringPtr("test"),
				ReadTimeout:       time.Second,
				ReadHeaderTimeout: 2 * time.Second,
				ShutdownTimeout:   3 * time.Second,
				Logger:            NewMockInfoer(nil),
				CancelHandler:     cancelHandler,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			originalCancelHandler := testCase.settings.CancelHandler

			testCase.settings.SetDefaults()

			require.NotNil(t, testCase.settings.CancelHandler)
			if originalCancelHandler != nil {
				assert.Equal(t,
					reflect.ValueOf(originalCancelHandler).Pointer(),
					reflect.ValueOf(testCase.expectedSettings.CancelHandler).Pointer())
			}
			// Remove function pointer before comparison
			testCase.expectedSettings.CancelHandler = nil
			testCase.settings.CancelHandler = nil

			assert.Equal(t, testCase.expectedSettings, testCase.settings)
		})
	}
}

func Test_Settings_Validate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings   Settings
		errMessage string
	}{
		"nil handler": {
			errMessage: "handler is nil",
		},
		"invalid settings": {
			settings: Settings{
				Handler: http.NewServeMux(),
				Address: stringPtr(":-1"),
			},
			errMessage: "listening address is not valid: address -1: invalid port",
		},
		"valid settings": {
			settings: Settings{
				Handler: http.NewServeMux(),
				Address: stringPtr(":0"),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.settings.Validate()

			if testCase.errMessage == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}
