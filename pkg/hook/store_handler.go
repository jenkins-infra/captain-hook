package hook

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/garethjevans/captain-hook/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type handler struct {
	InsecureRelay    bool
	client           *http.Client
	maxRetryDuration *time.Duration
	store            store.Store
	sender           sender
}

type sender interface {
	send(forwardURL string, bodyBytes []byte, header http.Header) error
}

func (h *handler) Handle(hook *Hook) error {
	// need to have a think about that the logic would be here.
	hookID, err := h.store.StoreHook(hook.ForwardURL, hook.Body, hook.Headers)
	if err != nil {
		return err
	}

	hook.ID = hookID

	if h.sender == nil {
		h.sender = h
	}
	// attempt to send
	err = h.sender.send(hook.ForwardURL, hook.Body, hook.Headers)
	if err != nil {
		// if failed, mark as failed with the error as the message
		err = h.store.Error(hookID, err.Error())
		if err != nil {
			return err
		}
	} else {
		// if success, mark as successful,
		err = h.store.Success(hookID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) send(forwardURL string, bodyBytes []byte, header http.Header) error {
	f := func() error {
		logrus.Debugf("relaying %s", string(bodyBytes))
		//g := hmac.NewGenerator("sha256", decodedHmac)
		//signature := g.HubSignature(bodyBytes)

		var httpClient *http.Client

		if h.client != nil {
			httpClient = h.client
		} else {
			if h.InsecureRelay {
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

		// does this need to be resigned?
		//req.Header.Add("X-Hub-Signature", signature)

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
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return errors.Errorf("%s not available, error was %s", req.URL.String(), resp.Status)
		}

		// And finally, if we haven't gotten any errors, just return nil because we're good.
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	// Try again after 2/4/8/... seconds if necessary, for up to 90 seconds, may take up to a minute to for the secret to replicate
	bo.InitialInterval = 2 * time.Second
	bo.MaxElapsedTime = 2 * (*h.maxRetryDuration)
	bo.Reset()

	return backoff.RetryNotify(f, bo, func(e error, t time.Duration) {
		logrus.Infof("webhook relaying failed: %s, backing off for %s", e, t)
	})
}
