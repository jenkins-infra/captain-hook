package hook

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha12 "github.com/garethjevans/captain-hook/pkg/api/captainhookio/v1alpha1"
	"github.com/garethjevans/captain-hook/pkg/client/clientset/versioned"
	"github.com/garethjevans/captain-hook/pkg/client/informers/externalversions"
	"github.com/garethjevans/captain-hook/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	resyncPeriod = time.Second * 30
)

type informer struct {
	handler *handler
}

func (i *informer) Start() error {
	logrus.Infof("getting in cluster config")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	logrus.Infof("creating client set")
	client, err := versioned.NewForConfig(config)
	if err != nil {
		return err
	}

	namespace, err := util.Namespace()
	if err != nil {
		return err
	}

	logrus.Infof("creating shared informer factory")
	factory := externalversions.NewSharedInformerFactoryWithOptions(client, resyncPeriod, externalversions.WithNamespace(namespace))
	informer := factory.Captainhook().V1alpha1().Hooks().Informer()

	stopper := make(chan struct{})

	defer close(stopper)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			h := obj.(*v1alpha12.Hook)
			logrus.Infof("Created hook in namespace %s, name %s at %s", h.ObjectMeta.Namespace, h.ObjectMeta.Name, h.ObjectMeta.CreationTimestamp)
		},
		DeleteFunc: func(obj interface{}) {
			h := obj.(*v1alpha12.Hook)
			logrus.Infof("Deleted hook in namespace %s, name %s", h.ObjectMeta.Namespace, h.ObjectMeta.Name)
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			h := newObj.(*v1alpha12.Hook)
			logrus.Infof("Updated hook in namespace %s, name %s at %s", h.ObjectMeta.Namespace, h.ObjectMeta.Name, h.ObjectMeta.CreationTimestamp)
			if h.Status.Phase == v1alpha12.HookPhaseSuccess {
				err := DeleteIfOld(client, h)
				if err != nil {
					logrus.Errorf("delete if old failed: %s", err.Error())
				}
			}
		},
	})
	informer.Run(stopper)

	return nil
}

func DeleteIfOld(client versioned.Interface, hook *v1alpha12.Hook) error {
	// if phase is successful
	if hook.Status.Phase == v1alpha12.HookPhaseSuccess {
		err := client.CaptainhookV1alpha1().Hooks(hook.ObjectMeta.Namespace).Delete(context.TODO(), hook.ObjectMeta.Name, v1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	// and age is more than period set

	// then delete
	return nil
}

func Retry(hook *v1alpha12.Hook) error {
	// if phase is failed or none

	// set phase to pending, increment attempt

	// attempt to send

	// mark as success if passed

	// mark as failed if errored

	// should probably add a timestamp to backoff until

	return nil
}
