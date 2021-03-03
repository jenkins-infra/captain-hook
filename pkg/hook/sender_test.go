package hook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSender(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var tests = []struct {
		name           string
		data           string
		responseBody   string
		responseStatus int
		error          bool
	}{
		{
			name:           "ok",
			data:           "testdata/push.json",
			responseBody:   "OK",
			responseStatus: 200,
		},
		{
			name:           "no content",
			data:           "testdata/push.json",
			responseBody:   "OK",
			responseStatus: 204,
		},
		{
			name:           "bad request",
			data:           "testdata/push.json",
			responseBody:   "Not OK",
			responseStatus: 400,
			error:          true,
		},
		{
			name:           "not found",
			data:           "testdata/push.json",
			responseBody:   "Not OK",
			responseStatus: 404,
			error:          true,
		},
		{
			name:           "internal server error",
			data:           "testdata/push.json",
			responseBody:   "Not OK",
			responseStatus: 500,
			error:          true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handlerFunc := func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, req.URL.String(), "/")

				rw.WriteHeader(test.responseStatus)

				if test.responseStatus != 204 {
					writeResult(rw, test.responseBody)
				}
			}

			server := httptest.NewServer(http.HandlerFunc(handlerFunc))
			// Close the server when test finishes
			defer server.Close()

			sender := sender{
				client: server.Client(),
			}

			data, err := ioutil.ReadFile(test.data)
			assert.NoError(t, err)

			buf := bytes.NewBuffer(data)

			header := make(http.Header)
			err = sender.send(server.URL, buf.Bytes(), header)

			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
