package hook

import (
	"time"

	"github.com/garethjevans/captain-hook/pkg/store"

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
	maxAgeInSeconds int
	client          versioned.Interface
	namespace       string
	sender          Sender
	store           store.Store
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
				logrus.Infof("Hook %s is success, checking age... %s", h.ObjectMeta.Name, h.Status.CompletedTimestamp)
				err := i.DeleteIfOld(h)
				if err != nil {
					logrus.Errorf("delete if old failed: %s", err.Error())
				}
			} else if h.Status.Phase == v1alpha12.HookPhaseFailed {
				now := v1.Now()
				if h.Status.NoRetryBefore.Before(&now) {
					err := i.Retry(h)
					if err != nil {
						logrus.Errorf("retry failed: %s", err.Error())
					}
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
		ts := v1.Now().Add(time.Second * time.Duration(-1*i.maxAgeInSeconds))
		logrus.Infof("checking if hook %s is older than %s", hook.ObjectMeta.Name, ts)
		if ts.After(hook.Status.CompletedTimestamp.Time) {
			// then delete
			err := i.store.Delete(hook.ObjectMeta.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *informer) Retry(hook *v1alpha12.Hook) error {
	// if phase is failed or none
	if hook.Status.Phase == v1alpha12.HookPhaseFailed {
		err := i.store.MarkForRetry(hook.ObjectMeta.Name)
		if err != nil {
			return err
		}

		// attempt to send
		err = i.sender.send(hook.Spec.ForwardURL, []byte(hook.Spec.Body), hook.Spec.Headers)

		if err != nil {
			// mark as failed if errored
			err = i.store.Error(hook.ObjectMeta.Name, err.Error())
			if err != nil {
				return err
			}
		} else {
			// mark as success if passed
			err = i.store.Success(hook.ObjectMeta.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
