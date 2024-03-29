package goservices

import (
	"testing"

	"github.com/qdm12/goservices/hooks"
	"github.com/stretchr/testify/assert"
)

func Test_RestarterSettings_setDefaults(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		originalSettings  RestarterSettings
		defaultedSettings RestarterSettings
	}{
		"empty settings": {
			defaultedSettings: RestarterSettings{
				Hooks: hooks.NewNoop(),
			},
		},
		"hooks already set": {
			originalSettings: RestarterSettings{
				Hooks: hooks.NewWithLog(nil),
			},
			defaultedSettings: RestarterSettings{
				Hooks: hooks.NewWithLog(nil),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testCase.originalSettings.setDefaults()
			assert.Equal(t, testCase.defaultedSettings, testCase.originalSettings)
		})
	}
}

func Test_RestarterSettings_validate(t *testing.T) {
	t.Parallel()

	dummyService := NewMockService(nil)

	testCases := map[string]struct {
		settings    RestarterSettings
		errSentinel error
		errMessage  string
	}{
		"missing service": {
			errSentinel: ErrNoService,
			errMessage:  "no service specified",
		},
		"minimal settings": {
			settings: RestarterSettings{
				Service: dummyService,
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.settings.validate()

			assert.ErrorIs(t, err, testCase.errSentinel)
			if testCase.errSentinel != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}
