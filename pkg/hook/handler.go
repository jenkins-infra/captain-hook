package hook

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jenkins-infra/captain-hook/pkg/store"

	"github.com/gorilla/mux"
	"github.com/jenkins-infra/captain-hook/pkg/version"
	"github.com/sirupsen/logrus"
)

const (
	HealthPath = "/health"
)

// Options struct containing all options.
type Options struct {
	Path       string
	Version    string
	ForwardURL string
	handler    *handler
	informer   *informer
}

// NewHook create a new hook handler.
func NewHook() (*Options, error) {
	logrus.Infof("creating new webhook listener")

	store := store.NewKubernetesStore()
	sender := NewSender()

	return &Options{
		Path:       os.Getenv("HOOK_PATH"),
		Version:    version.Version,
		ForwardURL: os.Getenv("FORWARD_URL"),
		handler: &handler{
			store:  store,
			sender: sender,
		},
		informer: &informer{
			maxAgeInSeconds: Atoi(os.Getenv("MAX_AGE_IN_SECONDS")),
			store:           store,
			sender:          sender,
		},
	}, nil
}

func (o *Options) Start() error {
	return o.informer.Start()
}

func (o *Options) Handle(mux *mux.Router) {
	logrus.Infof("Handling health on %s", HealthPath)
	mux.Handle(HealthPath, http.HandlerFunc(o.health))

	mux.Handle("/", http.HandlerFunc(o.defaultHandler))

	logrus.Infof("Handling hook on %s", o.Path)
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
	logrus.Infof("got incomming request")
	if r.Method != http.MethodPost {
		// liveness probe etc
		logrus.Info("invalid http method so returning index")
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

	logrus.Debugf("got hook body %s", string(bodyBytes))
	logrus.Debugf("got headers %s", r.Header)

	err = o.onGeneralHook(bodyBytes, r.Header)
	if err != nil {
		logrus.Errorf("failed to process webhook: %s", err)
		responseHTTPError(w, http.StatusInternalServerError, "500 Internal Server Error: %s", err.Error())
	}

	writeResult(w, "OK")
}

func (o *Options) onGeneralHook(bodyBytes []byte, headers http.Header) error {
	hook := Hook{
		ForwardURL: o.ForwardURL,
		Body:       bodyBytes,
		Headers:    headers,
	}

	err := o.handler.Handle(&hook)
	if err != nil {
		logrus.Errorf("failed to deliver webhook after %s", err)
		return err
	}

	logrus.Infof("webhook delivery ok for %s", hook.Name)

	return nil
}

func Atoi(in string) int {
	out, err := strconv.Atoi(in)
	if err != nil {
		panic(err)
	}
	return out
}
