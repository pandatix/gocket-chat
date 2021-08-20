package test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	gochat "github.com/pandatix/gocket-chat"
	"github.com/pandatix/gocket-chat/api"
)

// FakeHTTPClient is an implementation of HTTPClient that
// does nothing expect returning what you said it to.
// It's mainly made exported to mock everywhere (for test packages).
type FakeHTTPClient struct {
	Response *http.Response
	Err      error
}

func (f FakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return f.Response, f.Err
}

var _ gochat.HTTPClient = (*FakeHTTPClient)(nil)

// FakeReadCloser mocks an io.ReadCloser.
// It's mainly made exported to mock everywhere (for test packages).
type FakeReadCloser struct {
	data      []byte
	readIndex int64
}

func (f *FakeReadCloser) Read(p []byte) (n int, err error) {
	if f.readIndex >= int64(len(f.data)) {
		err = io.EOF
		return
	}

	n = copy(p, f.data[f.readIndex:])
	f.readIndex += int64(n)
	return
}

func (f *FakeReadCloser) Close() error {
	return nil
}

var _ = (io.ReadCloser)(&FakeReadCloser{})

// NewFakeReadCloser creates a new FakeReadCloser given a body. Only
// used for tests.
func NewFakeReadCloser(str string) *FakeReadCloser {
	return &FakeReadCloser{
		data: []byte(str),
	}
}

// NewTestRocketClient creates a new RocketClient given a body and an
// error to return. Only used for tests.
func NewTestRocketClient(body, url string, statusCode int, err error) *gochat.RocketClient {
	rc, _ := gochat.NewRocketClient(&FakeHTTPClient{
		Response: &http.Response{
			Body:       NewFakeReadCloser(body),
			StatusCode: statusCode,
		},
		Err: err,
	}, url, "", "")
	return rc
}

var errStrTypeOf = reflect.TypeOf(errors.New(""))

func CheckErr(err, expErr error, t *testing.T) {
	// Check err type
	typeErr := reflect.TypeOf(err)
	typeExpErr := reflect.TypeOf(expErr)
	if typeErr != typeExpErr {
		t.Fatalf("Failed to get expected error type: got \"%s\" instead of \"%s\".", typeErr, typeExpErr)
	}

	// Check Error content is not empty
	if err != nil && err.Error() == "" {
		t.Error("Error should not have an empty content.")
	}

	// Check if the error is generated using errors.New
	if typeErr == errStrTypeOf {
		if err.Error() != expErr.Error() {
			t.Errorf("Error message differs: got \"%s\" instead of \"%s\".", err, expErr)
		}
		return
	}

	switch err.(type) {
	case *url.Error:
		castedErr := err.(*url.Error)
		castedExpErr := expErr.(*url.Error)

		if castedErr.Op != castedExpErr.Op {
			t.Errorf("Failed to get expected Op: got \"%s\" instead of \"%s\".", castedErr.Op, castedExpErr.Op)
		}

		if castedErr.URL != castedExpErr.URL {
			t.Errorf("Failed to get expected URL: got \"%s\" instead of \"%s\".", castedErr.URL, castedExpErr.URL)
		}

	case *api.ErrUnexpectedStatus:
		castedErr := err.(*api.ErrUnexpectedStatus)
		castedExpErr := expErr.(*api.ErrUnexpectedStatus)

		if castedErr.StatusCode != castedExpErr.StatusCode {
			t.Errorf("Failed to get expected status code: got %d instead of %d.", castedErr.StatusCode, castedExpErr.StatusCode)
		}

	case *json.SyntaxError:
		castedErr := err.(*json.SyntaxError)
		castedExpErr := expErr.(*json.SyntaxError)

		if castedErr.Offset != castedExpErr.Offset {
			t.Errorf("Failed to get expected offset: got %d instead of %d.", castedErr.Offset, castedExpErr.Offset)
		}

	case *api.ErrCall:
		castedErr := err.(*api.ErrCall)
		castedExpErr := expErr.(*api.ErrCall)

		if castedErr.ErrorType != castedExpErr.ErrorType {
			t.Errorf("Failed to get expected error type: got \"%s\" instead of \"%s\".", castedErr.ErrorType, castedExpErr.ErrorType)
		}

	case nil:
		return

	default:
		t.Logf("\033[31mcheckErr Unsupported type: %s\033[0m\n", typeErr)
	}
}
