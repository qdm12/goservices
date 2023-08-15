package goservices

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/goservices/hooks"
	"github.com/stretchr/testify/assert"
)

func Test_GroupSettings_setDefaults(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		originalSettings  GroupSettings
		defaultedSettings GroupSettings
	}{
		"empty settings": {
			defaultedSettings: GroupSettings{
				Hooks: hooks.NewNoop(),
			},
		},
		"hooks already set": {
			originalSettings: GroupSettings{
				Hooks: hooks.NewWithLog(nil),
			},
			defaultedSettings: GroupSettings{
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

func Test_GroupSettings_validate(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	// Need to share the same service pointers so they are defined in the
	// parent test for all the subtests.
	dummyServiceOne := NewMockService(ctrl)
	dummyServiceOne.EXPECT().String().Return("dummy one").AnyTimes()
	dummyServiceTwo := NewMockService(ctrl)
	dummyServiceTwo.EXPECT().String().Return("dummy two").AnyTimes()

	testCases := map[string]struct {
		settings    GroupSettings
		errSentinel error
		errMessage  string
	}{
		"no service specified": {
			settings:    GroupSettings{},
			errSentinel: ErrNoService,
			errMessage:  "no service specified",
		},
		"nil service": {
			settings: GroupSettings{
				Services: []Service{nil},
			},
			errSentinel: ErrServiceIsNil,
			errMessage:  "service at index 0: service is nil",
		},
		"single service duplicated": {
			settings: GroupSettings{
				Services: []Service{dummyServiceOne, dummyServiceOne, dummyServiceOne},
			},
			errSentinel: ErrServicesNotUnique,
			errMessage:  "services are not unique: service dummy one is duplicated 3 times",
		},
		"multiple services duplicated": {
			settings: GroupSettings{
				Services: []Service{dummyServiceOne, dummyServiceOne, dummyServiceTwo, dummyServiceTwo},
			},
			errSentinel: ErrServicesNotUnique,
			errMessage: "services are not unique: services dummy one is duplicated twice " +
				"and dummy two is duplicated twice",
		},
		"success": {
			settings: GroupSettings{
				Services: []Service{dummyServiceOne},
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
