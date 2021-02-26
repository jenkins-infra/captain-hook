package hook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/garethjevans/captain-hook/pkg/store"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestWebhooks(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var tests = []struct {
		name             string
		event            string
		before           string
		multipleAttempts bool
		handlerFunc      func(rw http.ResponseWriter, req *http.Request)
	}{
		// push
		{
			name:             "should_relay",
			event:            "push",
			before:           "testdata/push.json",
			multipleAttempts: false,
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				// Test request parameters
				assert.Equal(t, req.URL.String(), "/")

				// Send response to be tested
				t.Logf("sending 'OK'")
				_, err := rw.Write([]byte(`OK`))
				assert.NoError(t, err)
			},
		},
		// any other error
		{
			name:             "error",
			event:            "push",
			before:           "testdata/push.json",
			multipleAttempts: true,
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				// Test request parameters
				assert.Equal(t, req.URL.String(), "/")

				// Send response to be tested
				rw.WriteHeader(500)
				t.Logf("sending 'not ok'")
				_, err := rw.Write([]byte(`not ok`))
				assert.NoError(t, err)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before, err := ioutil.ReadFile(test.before)
			if err != nil {
				t.Error(err)
			}

			buf := bytes.NewBuffer(before)
			r, _ := http.NewRequest("POST", "/", buf)
			r.Header.Set("X-GitHub-Event", test.event)
			r.Header.Set("X-GitHub-Delivery", "f2467dea-70d6-11e8-8955-3c83993e0aef")

			retryDuration := 5 * time.Second
			handler := Options{
				maxRetryDuration: &retryDuration,
				store:            store.NewLoggingStore(),
			}

			attempts := 0

			hf := func(rw http.ResponseWriter, req *http.Request) {
				attempts++
				test.handlerFunc(rw, req)
			}
			server := httptest.NewServer(http.HandlerFunc(hf))
			// Close the server when test finishes
			defer server.Close()

			handler.ForwardURL = server.URL
			handler.client = server.Client()

			w := NewFakeRespone(t)
			handler.handleWebHookRequests(w, r)

			assert.Equal(t, string(w.body), "OK")
			assert.Equal(t, test.multipleAttempts, attempts > 1, "expected multiple attempts %t, but got %d", test.multipleAttempts, attempts)
		})
	}
}

type FakeResponse struct {
	t       *testing.T
	headers http.Header
	body    []byte
	status  int
}

func NewFakeRespone(t *testing.T) *FakeResponse {
	return &FakeResponse{
		t:       t,
		headers: make(http.Header),
	}
}

func (r *FakeResponse) Header() http.Header {
	return r.headers
}

func (r *FakeResponse) Write(body []byte) (int, error) {
	r.body = body
	return len(body), nil
}

func (r *FakeResponse) WriteHeader(status int) {
	r.status = status
}
