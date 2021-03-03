package hook

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Sender interface {
	send(forwardURL string, bodyBytes []byte, header map[string][]string) error
}

type sender struct {
	client        *http.Client
	InsecureRelay bool
}

func NewSender() Sender {
	return &sender{
		InsecureRelay: os.Getenv("INSECURE_RELAY") == "true",
	}
}

func (s *sender) send(forwardURL string, bodyBytes []byte, header map[string][]string) error {
	logrus.Debugf("relaying %s", string(bodyBytes))

	var httpClient *http.Client

	if s.client != nil {
		httpClient = s.client
	} else {
		if s.InsecureRelay {
			// #nosec G402
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}

			httpClient = &http.Client{Transport: tr}
		} else {
			httpClient = &http.Client{}
		}
	}

	req, err := http.NewRequest("POST", forwardURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header = header

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	logrus.Infof("got resp code %d from url '%s'", resp.StatusCode, forwardURL)

	// If we got a 500, check if it's got the "repository not configured" string in the body. If so, we retry.
	if resp.StatusCode == 500 {
		respBody, err := ioutil.ReadAll(io.LimitReader(resp.Body, 10000000))
		if err != nil {
			return backoff.Permanent(errors.Wrap(err, "parsing resp.body"))
		}
		err = resp.Body.Close()
		if err != nil {
			return backoff.Permanent(errors.Wrap(err, "closing resp.body"))
		}
		logrus.Infof("got error respBody '%s'", string(respBody))
	}

	// If we got anything other than a 2xx, retry as well.
	// We're leaving this distinct from the "not configured" behavior in case we want to resurrect that later. (apb)
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return errors.Errorf("%s not available, error was %s", req.URL.String(), resp.Status)
	}

	// And finally, if we haven't gotten any errors, just return nil because we're good.
	return nil
}
