package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_SetDefaults(t *testing.T) {
	t.Parallel()

	errTest := fmt.Errorf("test error")

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
				OnStop:            func(ctx context.Context) error { return nil },
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
				OnStop:            func(ctx context.Context) error { return errTest },
			},
			expectedSettings: Settings{
				Name:              stringPtr("x"),
				Handler:           http.NewServeMux(),
				Address:           stringPtr("test"),
				ReadTimeout:       time.Second,
				ReadHeaderTimeout: 2 * time.Second,
				ShutdownTimeout:   3 * time.Second,
				Logger:            NewMockInfoer(nil),
				OnStop:            func(ctx context.Context) error { return errTest },
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testCase.settings.SetDefaults()

			expectedFuncResult := testCase.expectedSettings.OnStop(context.Background())
			actualFuncResult := testCase.settings.OnStop(context.Background())
			assert.Equal(t, expectedFuncResult, actualFuncResult)
			// Set the function to nil to be able to compare the structs
			testCase.settings.OnStop = nil
			testCase.expectedSettings.OnStop = nil
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
		testCase := testCase
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
