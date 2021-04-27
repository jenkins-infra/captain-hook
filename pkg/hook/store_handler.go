package hook

import (
	"github.com/jenkins-infra/captain-hook/pkg/store"
)

type handler struct {
	store  store.Store
	sender Sender
}

func (h *handler) Handle(hook *Hook) error {
	// need to have a think about that the logic would be here.
	hookName, err := h.store.StoreHook(hook.ForwardURL, hook.Body, hook.Headers)
	if err != nil {
		return err
	}

	hook.Name = hookName

	if h.sender == nil {
		h.sender = NewSender()
	}
	// attempt to send
	err = h.sender.send(hook.ForwardURL, hook.Body, hook.Headers)
	if err != nil {
		// if failed, mark as failed with the error as the message
		err = h.store.Error(hookName, err.Error())
		if err != nil {
			return err
		}
	} else {
		// if success, mark as successful,
		err = h.store.Success(hookName)
		if err != nil {
			return err
		}
	}

	return nil
}
