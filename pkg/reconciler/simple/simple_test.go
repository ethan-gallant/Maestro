package simple

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestConfigMapUpdate(t *testing.T) {
	// Initialize scheme
	s := runtime.NewScheme()
	s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.ConfigMap{})

	// Create a fake client with a ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "default",
		},
		Data: map[string]string{
			"key": "initial-value",
		},
	}

	k8sCli := fake.NewClientBuilder().WithScheme(s).WithObjects(configMap).Build()

	// Define a reconcile function that updates the ConfigMap
	reconcileFn := func(ctx context.Context, parent client.Object) (client.Object, error) {
		cm := &corev1.ConfigMap{}
		err := k8sCli.Get(ctx, types.NamespacedName{Name: parent.GetName(), Namespace: parent.GetNamespace()}, cm)
		if err != nil {
			return nil, err
		}
		cm.Data["key"] = "updated-value"
		return cm, nil
	}

	// Create a SimpleReconciler
	reconciler := Reconciler[client.Object, client.Object]{
		ReconcileFn: reconcileFn,
	}

	// Call the Reconcile method
	result, err := reconciler.Reconcile(context.Background(), k8sCli, configMap)
	assert.NoError(t, err)
	require.True(t, result.Requeue || result.RequeueAfter > 0)

	// Get the updated ConfigMap
	updatedConfigMap := &corev1.ConfigMap{}
	err = k8sCli.Get(context.Background(), types.NamespacedName{Name: "test-configmap", Namespace: "default"}, updatedConfigMap)
	assert.NoError(t, err)

	// Assert that the ConfigMap data was updated
	assert.Equal(t, "updated-value", updatedConfigMap.Data["key"])

	// Now ensure that the ConfigMap is not updated if the reconcile function returns an error
	// Call the Reconcile method
	result, err = reconciler.Reconcile(context.Background(), k8sCli, configMap)
	assert.NoError(t, err)
	require.False(t, result.Requeue || result.RequeueAfter > 0)

	// Ensure that the ConfigMap data was not updated
	compareConfigMap := &corev1.ConfigMap{}
	err = k8sCli.Get(context.Background(), types.NamespacedName{Name: "test-configmap", Namespace: "default"}, compareConfigMap)
	assert.NoError(t, err)
	assert.Equal(t, *updatedConfigMap, *compareConfigMap)

	// Ensure a failure in the reconcile function is propagated
	reconcileFn = func(ctx context.Context, parent client.Object) (client.Object, error) {
		return nil, assert.AnError
	}
	reconciler = Reconciler[client.Object, client.Object]{
		ReconcileFn: reconcileFn,
	}

	result, err = reconciler.Reconcile(context.Background(), k8sCli, configMap)
	assert.Error(t, err)
	require.False(t, result.Requeue || result.RequeueAfter > 0)
}
