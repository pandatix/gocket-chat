package gocketchat_test

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	gochat "github.com/pandatix/gocket-chat"
	"github.com/pandatix/gocket-chat/internal/test"
)

func TestNewRocketClient(t *testing.T) {
	t.Parallel()

	var tests = map[string]struct {
		Client               gochat.HTTPClient
		ExpectedRocketClient *gochat.RocketClient
		ExpectedErr          error
	}{
		"nil-client": {
			Client:               nil,
			ExpectedRocketClient: nil,
			ExpectedErr:          gochat.ErrNilClient,
		},
		"valid-client": {
			Client:               &test.FakeHTTPClient{},
			ExpectedRocketClient: &gochat.RocketClient{},
			ExpectedErr:          nil,
		},
	}

	for testname, tt := range tests {
		t.Run(testname, func(t *testing.T) {
			rc, err := gochat.NewRocketClient(tt.Client, "", "", "")

			// As you can't check the unexported fiels of a RocketClient,
			// check it's what was expected.
			if reflect.TypeOf(rc) != reflect.TypeOf(tt.ExpectedRocketClient) {
				t.Errorf("Failed to get expected RocketClient: got \"%v\" instead of \"%v\".", rc, tt.ExpectedErr)
			}

			if err != tt.ExpectedErr {
				t.Errorf("Failed to get expected error: got \"%v\" instead of \"%v\".", err, tt.ExpectedErr)
			}
		})
	}
}

func TestRocketClientDo(t *testing.T) {
	t.Parallel()

	invalidURL := "i%%v@l1d_Ur1"

	var tests = map[string]struct {
		RocketClient     *gochat.RocketClient
		Request          *http.Request
		ExpectedResponse *http.Response
		ExpectedErr      error
	}{
		"nil-request": {
			RocketClient:     test.NewTestRocketClient("", "", 0, nil),
			Request:          nil,
			ExpectedResponse: nil,
			ExpectedErr:      gochat.ErrNilRequest,
		},
		"invalid-url": {
			RocketClient: test.NewTestRocketClient("", invalidURL, 0, nil),
			Request: &http.Request{
				URL:    &url.URL{},
				Header: http.Header{},
			},
			ExpectedResponse: nil,
			ExpectedErr: &url.Error{
				Op:  "parse",
				URL: invalidURL,
				Err: url.EscapeError("%%v"),
			},
		},
		"valid-call": {
			RocketClient: test.NewTestRocketClient("payload", "", http.StatusOK, nil),
			Request: &http.Request{
				URL:    &url.URL{},
				Header: http.Header{},
			},
			ExpectedResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       test.NewFakeReadCloser("payload"),
			},
			ExpectedErr: nil,
		},
	}

	for testname, tt := range tests {
		t.Run(testname, func(t *testing.T) {
			res, err := tt.RocketClient.Do(tt.Request)

			if !reflect.DeepEqual(res, tt.ExpectedResponse) {
				t.Errorf("Failed to get expected response: got \"%v\" instead of \"%v\".", res, tt.ExpectedResponse)
			}

			test.CheckErr(err, tt.ExpectedErr, t)
		})
	}
}
