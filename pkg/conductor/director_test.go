package conductor

import (
	"context"
	"github.com/ethan-gallant/maestro/api"
	corev1 "k8s.io/api/core/v1"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MockReconciler[Parent client.Object] struct {
	Called bool
	Ctx    context.Context
	Client client.Client
	Parent Parent
}

func (m *MockReconciler[Parent]) Describe() api.Descriptor {
	return api.Descriptor{
		Name:        "MockReconciler",
		Description: "A mock reconciler for testing",
	}
}

func (m *MockReconciler[Parent]) Reconcile(ctx context.Context, c client.Client, parent Parent) (reconcile.Result, error) {
	m.Called = true
	m.Ctx = ctx
	m.Client = c
	m.Parent = parent
	return reconcile.Result{}, nil
}

var _ api.Reconciler[client.Object] = &MockReconciler[client.Object]{}

func TestDirector(t *testing.T) {
	ctx := context.Background()
	mockClient := fake.NewClientBuilder().Build()
	mockLogger := logr.Discard()
	mockParent := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	director := ForParent(mockParent).
		WithClient(mockClient).
		WithContext(ctx).
		WithLogger(mockLogger).
		Build()

	if director.client != mockClient || director.parent != mockParent {
		t.Errorf("NewDirectorFor did not correctly initialize Director")
	}

	mockReconciler := &MockReconciler[*corev1.Pod]{}
	_, err := director.Reconcile(mockReconciler)
	if err != nil {
		t.Errorf("Reconcile returned an unexpected error: %v", err)
	}
	if !mockReconciler.Called {
		t.Errorf("Reconcile did not call Reconciler's Reconcile method")
	}
	if mockReconciler.Ctx != director.ctx || mockReconciler.Client != director.client || mockReconciler.Parent != director.parent {
		t.Errorf("Reconcile did not pass the correct parameters to the Reconciler")
	}
}
