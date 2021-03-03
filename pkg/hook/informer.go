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
	handler         *handler
	maxAgeInSeconds int
	client          versioned.Interface
	namespace       string
}

func (i *informer) Start() error {
	if i.client == nil {
		logrus.Infof("getting in cluster config")
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}

		logrus.Infof("creating client set")
		i.client, err = versioned.NewForConfig(config)
		if err != nil {
			return err
		}

		i.namespace, err = util.Namespace()
		if err != nil {
			return err
		}
	}

	logrus.Infof("creating shared informer factory")
	factory := externalversions.NewSharedInformerFactoryWithOptions(i.client, resyncPeriod, externalversions.WithNamespace(i.namespace))
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
				err := i.DeleteIfOld(h)
				if err != nil {
					logrus.Errorf("delete if old failed: %s", err.Error())
				}
			}
		},
	})
	informer.Run(stopper)

	return nil
}

func (i *informer) DeleteIfOld(hook *v1alpha12.Hook) error {
	// if phase is successful
	if hook.Status.Phase == v1alpha12.HookPhaseSuccess {
		// and age is more than period set
		if hook.Status.CompletedTimestamp.After(v1.Now().Add(time.Second * time.Duration(i.maxAgeInSeconds))) {
			// then delete
			err := i.client.CaptainhookV1alpha1().Hooks(hook.ObjectMeta.Namespace).Delete(context.TODO(), hook.ObjectMeta.Name, v1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

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
