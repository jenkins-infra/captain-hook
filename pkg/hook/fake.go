package hook

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"
)

type fakeSender struct {
	fail bool
	Urls []string
}

func (f *fakeSender) send(forwardURL string, bodyBytes []byte, header map[string][]string) error {
	f.Urls = append(f.Urls, forwardURL)
	if f.fail {
		return errors.New("simulate a failure")
	}
	return nil
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
