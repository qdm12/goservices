package httpserver

import (
	"context"
	"net"
	"net/http"
	"regexp"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/qdm12/goservices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings       Settings
		expectedServer *Server
		errMessage     string
	}{
		"invalid settings": {
			errMessage: "validating settings: handler is nil",
		},
		"valid settings": {
			settings: Settings{
				Handler: http.NewServeMux(),
			},
			expectedServer: &Server{
				settings: Settings{
					Name:              stringPtr(""),
					Handler:           http.NewServeMux(),
					Address:           stringPtr(""),
					ShutdownTimeout:   3 * time.Second,
					ReadTimeout:       10 * time.Second,
					ReadHeaderTimeout: time.Second,
					Logger:            &noopLogger{},
				},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := New(testCase.settings)

			if testCase.errMessage == "" {
				assert.NoError(t, err)
				assert.NotNil(t, server.service)
				server.service = nil
			} else {
				assert.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.expectedServer, server)
		})
	}
}

func Test_Server_String(t *testing.T) {
	t.Parallel()

	server := &Server{
		settings: Settings{
			Name: stringPtr("test"),
		},
	}

	assert.Equal(t, "test http server", server.String())

	server.settings.Name = stringPtr("")
	assert.Equal(t, "http server", server.String())
}

func Test_Server_GetAddress(t *testing.T) {
	t.Parallel()

	server := &Server{
		listeningAddress: "x",
	}

	address := server.GetAddress()

	assert.Equal(t, "x", address)
}

func Test_Server_success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logger := NewMockInfoer(ctrl)
	logger.EXPECT().Info(newRegexMatcher("^test http server listening on 127.0.0.1:[1-9][0-9]{0,4}$"))

	server, err := New(Settings{
		Handler:         handler,
		Name:            stringPtr("test"),
		Address:         stringPtr("127.0.0.1:0"),
		ShutdownTimeout: 10 * time.Second,
		Logger:          logger,
	})
	require.NoError(t, err)

	runError, err := server.Start(context.Background())
	require.NoError(t, err)

	addressRegex := regexp.MustCompile(`^127.0.0.1:[1-9][0-9]{0,4}$`)
	address := server.GetAddress()
	assert.Regexp(t, addressRegex, address)

	client := &http.Client{
		Timeout: time.Second,
	}
	_, port, err := net.SplitHostPort(address)
	require.NoError(t, err)
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:"+port, nil)
	require.NoError(t, err)
	response, err := client.Do(request)
	require.NoError(t, err)
	_ = response.Body.Close()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	select {
	case err := <-runError:
		require.NoError(t, err)
	default:
	}

	err = server.Stop()
	require.NoError(t, err)
}

func Test_Server_startError(t *testing.T) {
	t.Parallel()

	server := &Server{
		settings: Settings{
			Address:         stringPtr("127.0.0.1:-1"),
			ShutdownTimeout: 10 * time.Second,
		},
	}

	serverService := goservices.NewRunWrapper("server", server.run)

	runtimeError, err := serverService.Start(context.Background())

	require.EqualError(t, err, "listen tcp: address -1: invalid port")
	assert.Nil(t, runtimeError)
}
