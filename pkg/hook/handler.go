package hook

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/garethjevans/captain-hook/pkg/version"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	HealthPath = "/health"
)

var (
	defaultMaxRetryDuration = 45 * time.Second
)

// Options struct containing all options.
type Options struct {
	Port             string
	Path             string
	Version          string
	ForwardURL       string
	client           *http.Client
	maxRetryDuration *time.Duration
}

// NewHook create a new hook handler.
func NewHook() (*Options, error) {
	logrus.Infof("creating new webhook listener")
	return &Options{
		Path:             os.Getenv("HOOK_PATH"),
		Port:             os.Getenv("HOOK_PORT"),
		ForwardURL:       os.Getenv("FORWARD_URL"),
		Version:          version.Version,
		maxRetryDuration: &defaultMaxRetryDuration,
	}, nil
}

func (o *Options) Handle(mux *mux.Router) {
	mux.Handle(HealthPath, http.HandlerFunc(o.health))

	mux.Handle("/", http.HandlerFunc(o.defaultHandler))
	mux.Handle(o.Path, http.HandlerFunc(o.handleWebHookRequests))
}

// health returns either HTTP 204 if the service is healthy, otherwise nothing ('cos it's dead).
func (o *Options) health(w http.ResponseWriter, r *http.Request) {
	logrus.Trace("Health check")
	w.WriteHeader(http.StatusNoContent)
}

func (o *Options) defaultHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == o.Path || strings.HasPrefix(path, o.Path+"/") {
		o.handleWebHookRequests(w, r)
		return
	}
	path = strings.TrimPrefix(path, "/")
	if path == "" || path == "index.html" {
		o.getIndex(w)
		return
	}
	http.Error(w, fmt.Sprintf("unknown path %s", path), 404)
}

// getIndex returns a simple home page.
func (o *Options) getIndex(w io.Writer) {
	logrus.Debug("GET index")
	message := "Captain Hook"

	_, err := w.Write([]byte(message))
	if err != nil {
		logrus.Debugf("failed to write the index: %v", err)
	}
}

func (o *Options) handleWebHookRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// liveness probe etc
		logrus.Debug("invalid http method so returning index")
		o.getIndex(w)
		return
	}

	bodyBytes, err := ioutil.ReadAll(io.LimitReader(r.Body, 10000000))
	if err != nil {
		logrus.Errorf("failed to Read Body: %s", err.Error())
		responseHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("500 Internal Server Error: Read Body: %s", err.Error()))
		return
	}

	err = r.Body.Close() // must close
	if err != nil {
		logrus.Errorf("failed to Close Body: %s", err.Error())
		responseHTTPError(w, http.StatusInternalServerError, fmt.Sprintf("500 Internal Server Error: Read Close: %s", err.Error()))
		return
	}

	logrus.Debugf("got hook %s", string(bodyBytes))

	githubDeliveryEvent := r.Header.Get("X-GitHub-Delivery")
	githubEventType := r.Header.Get("X-GitHub-Event")

	err = o.onGeneralHook(githubEventType, githubDeliveryEvent, bodyBytes)
	if err != nil {
		logrus.Errorf("failed to process webhook: %s", err)
		responseHTTPError(w, http.StatusInternalServerError, "500 Internal Server Error: %s", err.Error())
	}

	writeResult(w, "OK")
}

func (o *Options) onGeneralHook(githubEventType string, githubDeliveryEvent string, bodyBytes []byte) error {
	// Set a default max retry duration of 30 seconds if it's not set.
	if o.maxRetryDuration == nil {
		o.maxRetryDuration = &defaultMaxRetryDuration
	}

	logrus.Debugf("onGeneralHook - %s", githubEventType)
	//decodedHmac, err := base64.StdEncoding.DecodeString(ws.HMAC)
	//if err != nil {
	//	log.WithError(err).Errorf("unable to decode hmac")
	//}

	err := o.retryWebhookDelivery(o.ForwardURL, githubEventType, githubDeliveryEvent, bodyBytes)
	if err != nil {
		logrus.Errorf("failed to deliver webhook after %s, %s", o.maxRetryDuration, err)
		return err
	}

	logrus.Infof("webhook delivery ok for %s", githubDeliveryEvent)

	return nil
}

func (o *Options) retryWebhookDelivery(forwardURL string, githubEventType string, githubDeliveryEvent string, bodyBytes []byte) error {
	f := func() error {
		logrus.Debugf("relaying %s", string(bodyBytes))
		//g := hmac.NewGenerator("sha256", decodedHmac)
		//signature := g.HubSignature(bodyBytes)

		var httpClient *http.Client

		if o.client != nil {
			httpClient = o.client
		} else {
			//if useInsecureRelay {
			//	// #nosec G402
			//	tr := &http.Transport{
			//		TLSClientConfig: &tls.Config{
			//			InsecureSkipVerify: true,
			//		},
			//	}

			//	httpClient = &http.Client{Transport: tr}
			//} else {
			httpClient = &http.Client{}
			//}
		}

		req, err := http.NewRequest("POST", forwardURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Add("X-GitHub-Event", githubEventType)
		req.Header.Add("X-GitHub-Delivery", githubDeliveryEvent)
		//req.Header.Add("X-Hub-Signature", signature)

		resp, err := httpClient.Do(req)
		logrus.Infof("got resp code %d from url '%s'", resp.StatusCode, forwardURL)
		if err != nil {
			return err
		}

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
	bo.MaxElapsedTime = 2 * (*o.maxRetryDuration)
	bo.Reset()

	return backoff.RetryNotify(f, bo, func(e error, t time.Duration) {
		logrus.Infof("webhook relaying failed: %s, backing off for %s", e, t)
	})
}
