/*
Package facebook contains the processes, functions and data types
required to receive data from and make requests to  Facebook's API.
*/
package facebook

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/bobheadxi/calories/config"
)

var (
	// GraphAPI : Specifies host used for Facebook API requests
	GraphAPI = "https://graph.facebook.com"
)

// MessageHandler : Called when a new message is received
type MessageHandler func(Event, Sender, ReceivedMessage)

// PostbackHandler : Called when a postback is received
type PostbackHandler func(Event, Sender, Postback)

// API : The service that handles all callbacks from Facebook,
// sorts events and passes them to appropriate handlers
type API struct {
	PageID          string
	Token           string
	MessageHandler  MessageHandler
	PostbackHandler PostbackHandler
}

// New : Build a new instance of our Facebook API service
func New(config *config.EnvConfig) *API {
	return &API{
		PageID: config.PageID,
		Token:  config.Token,
	}
}

// Handler : Listens for all HTTP requests and decides what to do with them
func (api *API) Handler(rw http.ResponseWriter, req *http.Request) {
	switch method := req.Method; method {
	case "POST":
		api.handlePOST(rw, req)

	case "GET":
		// Verify request authentication, return OK if match
		query := req.URL.Query()
		verifyToken := query.Get("hub.verify_token")
		if verifyToken == api.Token {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte(query.Get("hub.challenge")))
		} else {
			rw.WriteHeader(http.StatusUnauthorized)
		}

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HandlePOST : Works on POST requests, finds events that need to be handled.
func (api *API) handlePOST(rw http.ResponseWriter, req *http.Request) {
	read, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}

	event := &EventStream{}
	err = json.Unmarshal(read, event)
	if err != nil {
		return
	}

	// Iterate through event entries
	for _, entry := range event.Entries {
		for _, message := range entry.Messaging {
			if message.Message != nil {
				// Handle Message event
				go api.MessageHandler(entry.Event, *message.Sender, *message.Message)
			} else if message.Postback != nil {
				// Handle Postback event
				go api.PostbackHandler(entry.Event, *message.Sender, *message.Postback)
			}
		}
	}
	rw.WriteHeader(http.StatusOK)
}

// makeRequest : Makes a request to Facebook API
func (api *API) makeRequest(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	query := req.URL.Query()
	query.Set("access_token", api.Token)
	req.URL.RawQuery = query.Encode()
	return http.DefaultClient.Do(req)
}
