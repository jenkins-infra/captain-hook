package hook

import (
	"time"

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
		},
	})
	informer.Run(stopper)

	return nil
}
