package store

import "net/http"

// Store interface to implement a storage strategy.
type Store interface {
	// StoreHook stores a webhook in the store.
	StoreHook(forwardURL string, body string, header http.Header) error
}
