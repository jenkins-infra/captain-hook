package store

import "github.com/google/uuid"

type FakeStore struct {
	MessageIDs        []string
	SuccessMessageIDs []string
	ErroredMessageIDs []string
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
