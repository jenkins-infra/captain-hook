package hook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jenkins-infra/captain-hook/pkg/store"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestWebhooks(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var tests = []struct {
		name        string
		event       string
		before      string
		handlerFunc func(rw http.ResponseWriter, req *http.Request)
	}{
		// push
		{
			name:   "should_relay",
			event:  "push",
			before: "testdata/push.json",
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
			name:   "error",
			event:  "push",
			before: "testdata/push.json",
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

			hf := func(rw http.ResponseWriter, req *http.Request) {
				test.handlerFunc(rw, req)
			}

			server := httptest.NewServer(http.HandlerFunc(hf))
			// Close the server when test finishes
			defer server.Close()

			o := Options{
				ForwardURL: server.URL,
				handler: &handler{
					store: &store.FakeStore{},
					sender: &sender{
						client:        server.Client(),
						InsecureRelay: false,
					},
				},
			}

			w := NewFakeRespone(t)
			o.handleWebHookRequests(w, r)

			assert.Equal(t, string(w.body), "OK")
		})
	}
}
