package store

import (
	"context"

	v1alpha12 "github.com/garethjevans/captain-hook/pkg/api/captainhookio/v1alpha1"
	"github.com/garethjevans/captain-hook/pkg/client/clientset/versioned/typed/captainhookio/v1alpha1"
	"github.com/garethjevans/captain-hook/pkg/util"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type kubernetesStore struct {
	config *rest.Config
}

func NewKubernetesStore() Store {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	return &kubernetesStore{config: config}
}

func (s *kubernetesStore) StoreHook(forwardURL string, body []byte, header map[string][]string) (string, error) {
	cs, namespace, err := s.configAndNamespace()
	if err != nil {
		return "", err
	}

	hook := v1alpha12.Hook{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "hook-",
		},
		Spec: v1alpha12.HookSpec{
			ForwardURL: forwardURL,
			Body:       string(body),
			Headers:    header,
		},
		Status: v1alpha12.HookStatus{
			Phase: v1alpha12.HookPhasePending,
		},
	}

	created, err := cs.Hooks(namespace).Create(context.TODO(), &hook, v1.CreateOptions{})
	if err != nil {
		return "", err
	}

	logrus.Debugf("persisted hook %+v", created)

	return created.ObjectMeta.Name, nil
}

func (s *kubernetesStore) Success(id string) error {
	cs, namespace, err := s.configAndNamespace()
	if err != nil {
		return err
	}

	hook, err := cs.Hooks(namespace).Get(context.TODO(), id, v1.GetOptions{})
	if err != nil {
		return err
	}

	hook.Status.Phase = v1alpha12.HookPhaseSuccess
	hook.Status.Message = ""

	_, err = cs.Hooks(namespace).Update(context.TODO(), hook, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *kubernetesStore) Error(id string, message string) error {
	cs, namespace, err := s.configAndNamespace()
	if err != nil {
		return err
	}

	hook, err := cs.Hooks(namespace).Get(context.TODO(), id, v1.GetOptions{})
	if err != nil {
		return err
	}

	hook.Status.Phase = v1alpha12.HookPhaseFailed
	hook.Status.Message = message

	_, err = cs.Hooks(namespace).Update(context.TODO(), hook, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *kubernetesStore) Delete(id string) error {
	cs, namespace, err := s.configAndNamespace()
	if err != nil {
		return err
	}

	return cs.Hooks(namespace).Delete(context.TODO(), id, v1.DeleteOptions{})
}

func (s *kubernetesStore) configAndNamespace() (*v1alpha1.CaptainhookV1alpha1Client, string, error) {
	cs, err := v1alpha1.NewForConfig(s.config)
	if err != nil {
		return nil, "", err
	}

	logrus.Debugf("got clientset %s", cs)

	namespace, err := util.Namespace()
	if err != nil {
		return nil, "", err
	}

	logrus.Debugf("got namespace %s", namespace)

	return cs, namespace, nil
}
