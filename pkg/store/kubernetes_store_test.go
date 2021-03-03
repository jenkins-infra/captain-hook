package store

import (
	"net/http"
	"testing"

	v1alpha12 "github.com/garethjevans/captain-hook/pkg/api/captainhookio/v1alpha1"
	"github.com/garethjevans/captain-hook/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
)

func TestKubernetesStore_StoreHook(t *testing.T) {
	f := fake.Clientset{}

	store := kubernetesStore{
		namespace: "dummy",
		client:    &f,
	}

	// return a hook with a generated hook name as the fake
	f.AddReactor("create", "hooks",
		func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
			hook := action.(clienttesting.UpdateAction).GetObject().(*v1alpha12.Hook)
			assert.Equal(t, hook.Name, "")
			assert.Equal(t, hook.GenerateName, "hook-")

			assert.Equal(t, hook.Spec.ForwardURL, "http://thing.com")
			assert.Equal(t, hook.Spec.Body, "OK")
			assert.Equal(t, hook.Spec.Headers, make(map[string][]string))

			assert.Equal(t, hook.Status.Phase, v1alpha12.HookPhasePending)

			hook.ObjectMeta.Name = "generatedHookName"
			return true, hook, nil
		})

	hookName, err := store.StoreHook("http://thing.com", []byte("OK"), make(http.Header))
	assert.NoError(t, err)
	assert.Equal(t, "generatedHookName", hookName)
	assert.Equal(t, 1, len(f.Actions()))
	assert.Equal(t, "create", f.Actions()[0].GetVerb())
	assert.Equal(t, "hooks", f.Actions()[0].GetResource().Resource)
	assert.Equal(t, "v1alpha1", f.Actions()[0].GetResource().Version)
	assert.Equal(t, "captainhook.io", f.Actions()[0].GetResource().Group)
}

func TestKubernetesStore_Success(t *testing.T) {
	f := fake.Clientset{}

	store := kubernetesStore{
		namespace: "dummy",
		client:    &f,
	}

	f.AddReactor("get", "hooks",
		func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, &v1alpha12.Hook{
				ObjectMeta: v1.ObjectMeta{
					Name: "generatedHookName",
				},
				Spec: v1alpha12.HookSpec{
					ForwardURL: "http://test.com",
					Body:       "body",
					Headers:    nil,
				},
				Status: v1alpha12.HookStatus{
					Phase:    v1alpha12.HookPhasePending,
					Attempts: 0,
				},
			}, nil
		})

	f.AddReactor("update", "hooks",
		func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
			hook := action.(clienttesting.UpdateAction).GetObject().(*v1alpha12.Hook)
			assert.Equal(t, hook.Name, "generatedHookName")

			assert.Equal(t, hook.Status.Phase, v1alpha12.HookPhaseSuccess)
			assert.Equal(t, hook.Status.Message, "")
			assert.NotNil(t, hook.Status.CompletedTimestamp)

			return true, hook, nil
		})

	err := store.Success("hookName")
	assert.NoError(t, err)

	assert.Equal(t, 2, len(f.Actions()))

	assert.Equal(t, "get", f.Actions()[0].GetVerb())
	assert.Equal(t, "hooks", f.Actions()[0].GetResource().Resource)
	assert.Equal(t, "v1alpha1", f.Actions()[0].GetResource().Version)
	assert.Equal(t, "captainhook.io", f.Actions()[0].GetResource().Group)

	assert.Equal(t, "update", f.Actions()[1].GetVerb())
	assert.Equal(t, "hooks", f.Actions()[1].GetResource().Resource)
	assert.Equal(t, "v1alpha1", f.Actions()[1].GetResource().Version)
	assert.Equal(t, "captainhook.io", f.Actions()[1].GetResource().Group)
}

func TestKubernetesStore_Error(t *testing.T) {
	f := fake.Clientset{}

	store := kubernetesStore{
		namespace: "dummy",
		client:    &f,
	}

	f.AddReactor("get", "hooks",
		func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, &v1alpha12.Hook{
				ObjectMeta: v1.ObjectMeta{
					Name: "generatedHookName",
				},
				Spec: v1alpha12.HookSpec{
					ForwardURL: "http://test.com",
					Body:       "body",
					Headers:    nil,
				},
				Status: v1alpha12.HookStatus{
					Phase:    v1alpha12.HookPhasePending,
					Attempts: 0,
				},
			}, nil
		})

	f.AddReactor("update", "hooks",
		func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
			hook := action.(clienttesting.UpdateAction).GetObject().(*v1alpha12.Hook)
			assert.Equal(t, hook.Name, "generatedHookName")

			assert.Equal(t, hook.Status.Phase, v1alpha12.HookPhaseFailed)
			assert.Equal(t, hook.Status.Message, "dummyMessage")

			return true, hook, nil
		})

	err := store.Error("hookName", "dummyMessage")
	assert.NoError(t, err)

	assert.Equal(t, 2, len(f.Actions()))

	assert.Equal(t, "get", f.Actions()[0].GetVerb())
	assert.Equal(t, "hooks", f.Actions()[0].GetResource().Resource)
	assert.Equal(t, "v1alpha1", f.Actions()[0].GetResource().Version)
	assert.Equal(t, "captainhook.io", f.Actions()[0].GetResource().Group)

	assert.Equal(t, "update", f.Actions()[1].GetVerb())
	assert.Equal(t, "hooks", f.Actions()[1].GetResource().Resource)
	assert.Equal(t, "v1alpha1", f.Actions()[1].GetResource().Version)
	assert.Equal(t, "captainhook.io", f.Actions()[1].GetResource().Group)
}
