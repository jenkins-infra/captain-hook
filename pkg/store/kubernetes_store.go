package store

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	v1alpha12 "github.com/garethjevans/captain-hook/pkg/api/captainhookio/v1alpha1"
	"github.com/garethjevans/captain-hook/pkg/client/clientset/versioned/typed/captainhookio/v1alpha1"
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
	cs, err := v1alpha1.NewForConfig(s.config)
	if err != nil {
		return "", err
	}
	logrus.Debugf("got clientset %s", cs)

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
			Status: v1alpha12.HookStatusTypePending,
		},
	}

	logrus.Debugf("persisting hook %+v", hook)
	namespace, err := s.namespace()
	if err != nil {
		return "", err
	}
	created, err := cs.Hooks(namespace).Create(context.TODO(), &hook, v1.CreateOptions{})
	if err != nil {
		return "", err
	}
	logrus.Debugf("persisted hook %+v", created)

	return created.ObjectMeta.Name, nil
}

func (s *kubernetesStore) Success(id string) error {
	cs, err := v1alpha1.NewForConfig(s.config)
	if err != nil {
		return err
	}
	logrus.Debugf("got clientset %s", cs)

	namespace, err := s.namespace()
	if err != nil {
		return err
	}
	hook, err := cs.Hooks(namespace).Get(context.TODO(), id, v1.GetOptions{})
	if err != nil {
		return err
	}

	hook.Status.Status = v1alpha12.HookStatusTypeSuccess
	hook.Status.Message = ""

	_, err = cs.Hooks(namespace).Update(context.TODO(), hook, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *kubernetesStore) Error(id string, message string) error {
	cs, err := v1alpha1.NewForConfig(s.config)
	if err != nil {
		return err
	}
	logrus.Debugf("got clientset %s", cs)

	namespace, err := s.namespace()
	if err != nil {
		return err
	}
	hook, err := cs.Hooks(namespace).Get(context.TODO(), id, v1.GetOptions{})
	if err != nil {
		return err
	}

	hook.Status.Status = v1alpha12.HookStatusTypeFailed
	hook.Status.Message = message

	_, err = cs.Hooks(namespace).Update(context.TODO(), hook, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *kubernetesStore) namespace() (string, error) {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, nil
	}
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
		return "", err
	}
	return "", nil
}
