package hook

import (
	"net/http"

	"github.com/pkg/errors"
)

type fakeSender struct {
	fail bool
	Urls []string
}

func (f *fakeSender) send(forwardURL string, bodyBytes []byte, header http.Header) error {
	f.Urls = append(f.Urls, forwardURL)
	if f.fail {
		return errors.New("simulate a failure")
	}
	return nil
}
