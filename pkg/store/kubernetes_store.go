package store

import (
	"context"
	"net/http"

	v1alpha12 "github.com/garethjevans/captain-hook/pkg/api/captainhookio/v1alpha1"
	"github.com/garethjevans/captain-hook/pkg/client/clientset/versioned/typed/captainhookio/v1alpha1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type kubernetesStore struct {
	config *rest.Config
	namespace string
}

func NewKubernetesStore() Store {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	return &kubernetesStore{config: config}
}

func (s *kubernetesStore) StoreHook(forwardURL string, body string, header http.Header) error {
	cs, err := v1alpha1.NewForConfig(s.config)
	if err != nil {
		return err
	}
	logrus.Debugf("got clientset %s", cs)

	hook := v1alpha12.Hook{
		ObjectMeta: v1.ObjectMeta{},
		Spec: v1alpha12.HookSpec{
			ForwardURL: forwardURL,
			Body:       body,
			Headers:    header,
		},
		Status: v1alpha12.HookStatus{
			Status: v1alpha12.HookStatusTypePending,
		},
	}

	logrus.Debugf("persisting hook %+v", hook)
	created, err := cs.Hooks(s.namespace).Create(context.TODO(), &hook, v1.CreateOptions{})
	if err != nil {
		return err
	}
	logrus.Debugf("persisted hook %+v", created)

	return nil
}
