package hook

import v1alpha12 "github.com/garethjevans/captain-hook/pkg/api/captainhookio/v1alpha1"

// Hook struct to hold everything related to a hook.
type Hook struct {
	Name       string
	Namespace  string
	ForwardURL string
	Headers    map[string][]string
	Body       []byte
	// Status? State?

}

// FromV1Alpha1Hook converts from a v1alpha1.Hook to a Hook.
func FromV1Alpha1Hook(h *v1alpha12.Hook) Hook {
	return Hook{
		Name:       h.Name,
		Namespace:  h.Namespace,
		ForwardURL: h.Spec.ForwardURL,
		Headers:    h.Spec.Headers,
		Body:       []byte(h.Spec.Body),
	}
}
