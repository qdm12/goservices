package goservices

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_serviceError(t *testing.T) {
	t.Parallel()
	errTest := errors.New("test error")

	testCases := map[string]struct {
		serviceError serviceError
		errString    string
		errUnwrapped error
		panicValue   string
	}{
		"no err set panics": {
			serviceError: serviceError{
				format:      errorFormatCrash,
				serviceName: "A",
			},
			panicValue: "cannot have nil error in serviceError",
		},
		"error set": {
			serviceError: serviceError{
				format:      errorFormatCrash,
				serviceName: "A",
				err:         errTest,
			},
			errString:    "A crashed: test error",
			errUnwrapped: errTest,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if testCase.panicValue != "" {
				assert.PanicsWithValue(t, testCase.panicValue, func() {
					_ = testCase.serviceError.Error()
				})
				return
			}

			assert.ErrorIs(t, testCase.serviceError, testCase.errUnwrapped)
			assert.EqualError(t, testCase.serviceError, testCase.errString)
		})
	}
}

func Test_addStopError(t *testing.T) {
	t.Parallel()

	errTest := errors.New("test error")
	errTest2 := errors.New("test error 2")

	testCases := map[string]struct {
		collected           error
		serviceName         string
		newErr              error
		newCollectedErrors  []error
		newCollectedMessage string
	}{
		"all_nils": {},
		"collected_nil_new_error": {
			collected:           fmt.Errorf("stopping A: %w", errTest),
			newCollectedErrors:  []error{errTest},
			newCollectedMessage: "stopping A: test error",
		},
		"nil_collected_new_error": {
			serviceName:         "A",
			newErr:              errTest,
			newCollectedErrors:  []error{errTest},
			newCollectedMessage: "stopping A: test error",
		},
		"collected_new_error": {
			serviceName:         "B",
			collected:           fmt.Errorf("stopping A: %w", errTest),
			newErr:              errTest2,
			newCollectedErrors:  []error{errTest, errTest2},
			newCollectedMessage: "stopping A: test error; stopping B: test error 2",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			newCollected := addStopError(testCase.collected, testCase.serviceName,
				testCase.newErr)

			for _, err := range testCase.newCollectedErrors {
				assert.ErrorIs(t, newCollected, err)
			}
			if testCase.newCollectedMessage != "" {
				assert.EqualError(t, newCollected, testCase.newCollectedMessage)
			}
		})
	}
}
