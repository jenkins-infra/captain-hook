package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// Hook represents a webhook.
type Hook struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   HookSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status HookStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// HookSpec is the specification of a Hook.
type HookSpec struct {
	ForwardURL string              `json:"forwardURL" protobuf:"bytes,1,opt,name=forwardURL"`
	Body       string              `json:"body" protobuf:"bytes,2,opt,name=body"`
	Headers    map[string][]string `json:"headers,omitempty" protobuf:"bytes,3,opt,name=headers"`
}

// HookStatus is the status for a Hook resource.
type HookStatus struct {
	Status   HookStatusType `json:"status,omitempty" protobuf:"bytes,1,opt,name=status"`
	Attempts int            `json:"attempts,omitempty" protobuf:"bytes,2,opt,name=attempts"`
	Message  string         `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
}

// HookStatusType is the status of a hook; usually success or failed at completion.
type HookStatusType string

const (
	// HookStatusTypeNone an hook step has not started yet.
	HookStatusTypeNone HookStatusType = ""
	// HookStatusTypePending the hook currently being relayed.
	HookStatusTypePending HookStatusType = "Pending"
	// HookStatusTypeStatus the hook has been relayed.
	HookStatusTypeSuccess HookStatusType = "Success"
	// ReleaseStatusTypeFailed the hook has failed to be relayed.
	HookStatusTypeFailed HookStatusType = "Failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HookList is a list of TypeMeta resources.
type HookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Hook `json:"items"`
}
