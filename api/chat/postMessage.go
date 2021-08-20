package chat

import (
	"bytes"
	"encoding/json"
	"net/http"

	gochat "github.com/pandatix/gocket-chat"
	"github.com/pandatix/gocket-chat/api"
)

// PostMessageParams combines all the parameters for an
// /api/v1/chat.postMessage call.
type PostMessageParams struct {
	Channel      string         `json:"channel"`
	Text         *string        `json:"text,omitempty"`
	Alias        *string        `json:"alias,omitempty"`
	Emoji        *string        `json:"emoji,omitempty"`
	Avatar       *string        `json:"avatar,omitempty"`
	Attachements *[]Attachement `json:"attachements,omitempty"`
}

// Attachement is a sub-PostMessageParams struct.
type Attachement struct {
	Color             string  `json:"color"`
	Text              string  `json:"text"`
	TS                string  `json:"ts"`
	ThumbURL          string  `json:"thumb_url"`
	MessageLink       string  `json:"message_link"`
	Collapsed         bool    `json:"collapsed"`
	AuthorName        string  `json:"author_name"`
	AuthorLink        string  `json:"author_link"`
	AuthorIcon        string  `json:"author_icon"`
	Title             string  `json:"title"`
	TitleLink         string  `json:"title_link"`
	TitleLinkDownload bool    `json:"title_link_download"`
	ImageURL          string  `json:"image_url"`
	AudioURL          string  `json:"audio_url"`
	VideoURL          string  `json:"video_url"`
	Fields            []Field `json:"fields"`
}

// Field is a sub-Attachement struct.
type Field struct {
	Short *bool  `json:"short,omitempty"`
	Title string `json:"title"`
	Value string `json:"value"`
}

// PostMessageResponse combines all the core data returned
// by a call to the /api/v1/chat.postMessage endpoint.
type PostMessageResponse struct {
	Ts      int64   `json:"ts"`
	Channel string  `json:"channel"`
	Message Message `json:"message"`
}

// Message is a sub-PostMessageResponse struct.
// Does not support md field while it does not seems legit.
type Message struct {
	Alias       string        `json:"alias"`
	Msg         string        `json:"msg"`
	Attachments []Attachement `json:"attachments"`
	Parseurls   bool          `json:"parseUrls"`
	Groupable   bool          `json:"groupable"`
	Ts          string        `json:"ts"`
	U           U             `json:"u"`
	Rid         string        `json:"rid"`
	Urls        *[]URL        `json:"urls"`
	Mentions    *[]Mention    `json:"mentions"`
	Channels    []Channel     `json:"channels"`
	Updatedat   string        `json:"_updatedAt"`
	ID          string        `json:"_id"`
}

// U is a sub-Message struct.
type U struct {
	ID       string `json:"_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// URL is a sub-Message struct.
type URL struct {
	URL string `json:"url"`
}

// Mention is a sub-Message struct.
type Mention struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Type     string `json:"type"`
}

// Channel is a sub-Message struct.
type Channel struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
}

// PostMessage posts a message to a channel, given the parameters.
func PostMessage(rc *gochat.RocketClient, pmp PostMessageParams) (*PostMessageResponse, error) {
	if rc == nil {
		return nil, gochat.ErrNilRocketClient
	}

	// Marshal the PostMessageParams and build the request
	j, _ := json.Marshal(pmp)
	buf := bytes.NewBuffer(j)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/chat.postMessage", buf)

	// Issue the request
	res, err := rc.Do(req)
	if err != nil {
		return nil, err
	}

	// Check the status code
	if res.StatusCode != http.StatusOK {
		return nil, &api.ErrUnexpectedStatus{StatusCode: res.StatusCode}
	}

	// Decode the response
	type Response struct {
		Success   bool     `json:"success"`
		Ts        *int64   `json:"ts"`
		Channel   *string  `json:"channel"`
		Message   *Message `json:"message"`
		Error     *string  `json:"error"`
		ErrorType *string  `json:"errorType"`
	}
	var resp Response
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	// Check everything went fine
	if resp.ErrorType != nil {
		return nil, &api.ErrCall{ErrorType: *resp.ErrorType}
	}
	if !resp.Success {
		return nil, api.ErrFailureStatus
	}

	return &PostMessageResponse{
		Ts:      *resp.Ts,
		Channel: *resp.Channel,
		Message: *resp.Message,
	}, nil
}
