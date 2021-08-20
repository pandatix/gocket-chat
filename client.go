package gocketchat

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	// ErrNilClient is an error meaning the given client is nil.
	ErrNilClient = errors.New("given client is nil")
	// ErrNilRocketClient is an error meaning the given RocketClient is nil.
	ErrNilRocketClient = errors.New("given rocketclient is nil")
	// ErrNilRequest is an error meaning the given Request is nil.
	ErrNilRequest = errors.New("given request is nil")
)

// HTTPClient defines what a client has to implement in order to
// be a RocketClient sub-client.
// It enables to give a *http.Client when creating a RocketClient.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Ensure http.Client matches HTTPClient.
var _ HTTPClient = (*http.Client)(nil)

// RocketClient works as a middleware of a HTTPClient, saving the RocketChat
// instance URL, your token and user ID.
// Your calls to the API are achieved through it.
type RocketClient struct {
	client HTTPClient
	token  string
	userID string
	url    string
}

// Ensure RocketClient matches the HTTPClient to be able to chain multiple
// clients, as middlewares.
var _ HTTPClient = (*RocketClient)(nil)

// NewRocketClient creates a RocketClient, given the sub-client, RocketChat's
// URL, your token and user ID.
// Returns an error if the client is nil.
func NewRocketClient(client HTTPClient, url, token, userID string) (*RocketClient, error) {
	if client == nil {
		return nil, ErrNilClient
	}

	return &RocketClient{
		client: client,
		token:  token,
		userID: userID,
		url:    url,
	}, nil
}

// Do set up the auth and content type headers, adding the RocketChat's
// URL and returning a *Response.
func (rc RocketClient) Do(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Set auth headers
	req.Header.Add("X-Auth-Token", rc.token)
	req.Header.Add("X-User-Id", rc.userID)
	req.Header.Add("Content-type", "application/json")

	// Add RocketChat url
	url, err := url.Parse(rc.url + req.URL.String())
	if err != nil {
		return nil, err
	}
	req.URL = url

	// Issue the request
	return rc.client.Do(req)
}
