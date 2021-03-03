package hook

import (
	"testing"

	"github.com/garethjevans/captain-hook/pkg/store"
	"github.com/stretchr/testify/assert"
)

func TestHandle_Success(t *testing.T) {
	s := store.FakeStore{}
	sender := fakeSender{}
	h := handler{
		store:  &s,
		sender: &sender,
	}

	hook := Hook{
		ForwardURL: "http://example.com",
	}

	err := h.Handle(&hook)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(s.MessageIDs))
	assert.Equal(t, 1, len(s.SuccessMessageIDs))
	assert.Equal(t, 0, len(s.ErroredMessageIDs))
}

func TestHandle_Error(t *testing.T) {
	s := store.FakeStore{}
	sender := fakeSender{fail: true}
	h := handler{
		store:  &s,
		sender: &sender,
	}

	hook := Hook{
		ForwardURL: "http://example.com",
	}

	err := h.Handle(&hook)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(s.MessageIDs))
	assert.Equal(t, 0, len(s.SuccessMessageIDs))
	assert.Equal(t, 1, len(s.ErroredMessageIDs))
}
