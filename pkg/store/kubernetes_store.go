package store

import (
	"context"
	"time"

	"github.com/jenkins-infra/captain-hook/pkg/client/clientset/versioned"

	v1alpha12 "github.com/jenkins-infra/captain-hook/pkg/api/captainhookio/v1alpha1"
	"github.com/jenkins-infra/captain-hook/pkg/util"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ Store = &kubernetesStore{}

type kubernetesStore struct {
	namespace string
	client    versioned.Interface
}

func NewKubernetesStore() Store {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	client, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	namespace, err := util.Namespace()
	if err != nil {
		panic(err)
	}

	return &kubernetesStore{client: client, namespace: namespace}
}

func (s *kubernetesStore) StoreHook(forwardURL string, body []byte, header map[string][]string) (string, error) {
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

	created, err := s.client.CaptainhookV1alpha1().Hooks(s.namespace).Create(context.TODO(), &hook, v1.CreateOptions{})
	if err != nil {
		return "", err
	}

	logrus.Debugf("persisted hook %+v", created)

	return created.ObjectMeta.Name, nil
}

func (s *kubernetesStore) Success(id string) error {
	hook, err := s.client.CaptainhookV1alpha1().Hooks(s.namespace).Get(context.TODO(), id, v1.GetOptions{})
	if err != nil {
		return err
	}

	hook.Status.Phase = v1alpha12.HookPhaseSuccess
	hook.Status.Message = ""
	now := v1.Now()
	hook.Status.CompletedTimestamp = &now

	_, err = s.client.CaptainhookV1alpha1().Hooks(s.namespace).Update(context.TODO(), hook, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *kubernetesStore) Error(id string, message string) error {
	hook, err := s.client.CaptainhookV1alpha1().Hooks(s.namespace).Get(context.TODO(), id, v1.GetOptions{})
	if err != nil {
		return err
	}

	hook.Status.Phase = v1alpha12.HookPhaseFailed
	hook.Status.Message = message

	// FIXME need to add the correct time here
	retry := v1.NewTime(time.Now().Add(time.Minute * 1))
	hook.Status.NoRetryBefore = &retry

	_, err = s.client.CaptainhookV1alpha1().Hooks(s.namespace).Update(context.TODO(), hook, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *kubernetesStore) Delete(id string) error {
	return s.client.CaptainhookV1alpha1().Hooks(s.namespace).Delete(context.TODO(), id, v1.DeleteOptions{})
}

func (s *kubernetesStore) MarkForRetry(id string) error {
	hook, err := s.client.CaptainhookV1alpha1().Hooks(s.namespace).Get(context.TODO(), id, v1.GetOptions{})
	if err != nil {
		return err
	}

	hook.Status.Phase = v1alpha12.HookPhasePending
	hook.Status.Message = ""
	hook.Status.Attempts = hook.Status.Attempts + 1
	hook.Status.NoRetryBefore = nil

	_, err = s.client.CaptainhookV1alpha1().Hooks(s.namespace).Update(context.TODO(), hook, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
