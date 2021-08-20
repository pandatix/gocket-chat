package chat_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	gochat "github.com/pandatix/gocket-chat"
	"github.com/pandatix/gocket-chat/api"
	"github.com/pandatix/gocket-chat/api/chat"
	"github.com/pandatix/gocket-chat/internal/test"
)

var errFake = errors.New("fake err")

func TestPostMessage(t *testing.T) {
	t.Parallel()

	validPayload := `{"ts":1628462781310,"channel":"#general","message":{"alias":"","msg":"This is a test!","attachments":[],"parseUrls":true,"groupable":false,"ts":"2021-08-08T22:46:21.276Z","u":{"_id":"q9SRsnxqsu7PSejrL","username":"lucas.tesson","name":"Lucas TESSON"},"rid":"GENERAL","urls":[],"mentions":[],"channels":[],"md":[{"type":"PARAGRAPH","value":[{"type":"PLAIN_TEXT","value":"This is a test!"}]}],"_updatedAt":"2021-08-08T22:46:21.288Z","_id":"WehP8L8upTtL9fqNw"},"success":true}`

	var tests = map[string]struct {
		RocketClient     *gochat.RocketClient
		Pmp              chat.PostMessageParams
		ExpectedResponse *chat.PostMessageResponse
		ExpectedErr      error
	}{
		"nil-rocketclient": {
			RocketClient:     nil,
			Pmp:              chat.PostMessageParams{},
			ExpectedResponse: nil,
			ExpectedErr:      gochat.ErrNilRocketClient,
		},
		"failing-rocketclient": {
			RocketClient:     test.NewTestRocketClient("", "", 0, errFake),
			Pmp:              chat.PostMessageParams{},
			ExpectedResponse: nil,
			ExpectedErr:      errFake,
		},
		"unexpected-status": {
			RocketClient:     test.NewTestRocketClient("", "", 0, nil),
			Pmp:              chat.PostMessageParams{},
			ExpectedResponse: nil,
			ExpectedErr:      &api.ErrUnexpectedStatus{StatusCode: 0},
		},
		"empty-body-response": {
			RocketClient:     test.NewTestRocketClient("", "", http.StatusOK, nil),
			Pmp:              chat.PostMessageParams{},
			ExpectedResponse: nil,
			ExpectedErr:      io.EOF,
		},
		"invalid-json-response": {
			RocketClient:     test.NewTestRocketClient("{[}]", "", http.StatusOK, nil),
			Pmp:              chat.PostMessageParams{},
			ExpectedResponse: nil,
			ExpectedErr:      &json.SyntaxError{Offset: 2},
		},
		"error-call": {
			RocketClient:     test.NewTestRocketClient(`{"success":true,"errorType":"type"}`, "", http.StatusOK, nil),
			Pmp:              chat.PostMessageParams{},
			ExpectedResponse: nil,
			ExpectedErr:      &api.ErrCall{ErrorType: "type"},
		},
		"failure-status": {
			RocketClient:     test.NewTestRocketClient(`{"success":false}`, "", http.StatusOK, nil),
			Pmp:              chat.PostMessageParams{},
			ExpectedResponse: nil,
			ExpectedErr:      api.ErrFailureStatus,
		},
		"error-call-before-failure-status": {
			RocketClient:     test.NewTestRocketClient(`{"success":false,"errorType":"type"}`, "", http.StatusOK, nil),
			Pmp:              chat.PostMessageParams{},
			ExpectedResponse: nil,
			ExpectedErr:      &api.ErrCall{ErrorType: "type"},
		},
		"valid-call": {
			RocketClient: test.NewTestRocketClient(validPayload, "", http.StatusOK, nil),
			Pmp:          chat.PostMessageParams{},
			ExpectedResponse: &chat.PostMessageResponse{
				Ts:      1628462781310,
				Channel: "#general",
				Message: chat.Message{
					Alias:       "",
					Msg:         "This is a test!",
					Attachments: []chat.Attachement{},
					Parseurls:   true,
					Groupable:   false,
					Ts:          "2021-08-08T22:46:21.276Z",
					U: chat.U{
						ID:       "q9SRsnxqsu7PSejrL",
						Username: "lucas.tesson",
						Name:     "Lucas TESSON",
					},
					Rid:       "GENERAL",
					Urls:      &[]chat.URL{},
					Mentions:  &[]chat.Mention{},
					Channels:  []chat.Channel{},
					Updatedat: "2021-08-08T22:46:21.288Z",
					ID:        "WehP8L8upTtL9fqNw",
				},
			},
			ExpectedErr: nil,
		},
	}

	for testname, tt := range tests {
		t.Run(testname, func(t *testing.T) {
			resp, err := chat.PostMessage(tt.RocketClient, tt.Pmp)

			if !reflect.DeepEqual(resp, tt.ExpectedResponse) {
				t.Errorf("Failed to get expected response: got \"%v\" instead of \"%v\".", resp, tt.ExpectedResponse)
			}

			test.CheckErr(err, tt.ExpectedErr, t)
		})
	}
}
