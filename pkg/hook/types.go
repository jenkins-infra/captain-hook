package hook

// Hook struct to hold everything related to a hook.
type Hook struct {
	ID         string
	ForwardURL string
	Headers    map[string][]string
	Body       []byte
	// Status? State?

}
