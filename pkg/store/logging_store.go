package store

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type loggingStore struct {
}

func NewLoggingStore() Store {
	return &loggingStore{}
}

func (s *loggingStore) StoreHook(forwardURL string, body string, header http.Header) error {
	logrus.Debugf("storing hook to %s", forwardURL)

	return nil
}
