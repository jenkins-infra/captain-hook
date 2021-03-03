package store

import "github.com/google/uuid"

var _ Store = &FakeStore{}

type FakeStore struct {
	MessageIDs        []string
	SuccessMessageIDs []string
	ErroredMessageIDs []string
	DeletedMessageIDs []string
	RetryMessageIDs   []string
}

func (s *FakeStore) StoreHook(forwardURL string, body []byte, header map[string][]string) (string, error) {
	messageID := uuid.New().String()
	s.MessageIDs = append(s.MessageIDs, messageID)
	return messageID, nil
}

func (s *FakeStore) Success(id string) error {
	s.SuccessMessageIDs = append(s.SuccessMessageIDs, id)
	return nil
}

func (s *FakeStore) Error(id string, message string) error {
	s.ErroredMessageIDs = append(s.ErroredMessageIDs, id)
	return nil
}

func (s *FakeStore) Delete(id string) error {
	s.DeletedMessageIDs = append(s.DeletedMessageIDs, id)
	return nil
}

func (s *FakeStore) MarkForRetry(id string) error {
	s.RetryMessageIDs = append(s.RetryMessageIDs, id)
	return nil
}
