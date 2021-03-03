package store

// Store interface to implement a storage strategy.
type Store interface {
	// StoreHook stores a webhook in the store.
	StoreHook(forwardURL string, body []byte, headers map[string][]string) (string, error)

	// Success marks a hook as successful.
	Success(id string) error

	// Marks a hook as error, with the error message.
	Error(id string, message string) error

	// Deletes a hook from the store.
	Delete(id string) error

	// Updates a hook to state it has been reattempted.
	MarkForRetry(id string) error
}
