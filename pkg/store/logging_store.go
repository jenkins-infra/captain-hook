package store

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type loggingStore struct {
}

func NewLoggingStore() Store {
	return &loggingStore{}
}

func (s *loggingStore) StoreHook(forwardURL string, body []byte, header map[string][]string) (string, error) {
	logrus.Debugf("storing hook to %s", forwardURL)
	uuid := uuid.New()
	return uuid.String(), nil
}

func (s *loggingStore) Success(id string) error {
	logrus.Infof("hook is successful: %s", id)
	return nil
}

func (s *loggingStore) Error(id string, message string) error {
	logrus.Errorf("hook is errored: %s, %s", message, id)
	return nil
}
