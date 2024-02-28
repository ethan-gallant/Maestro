package api

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler[Parent client.Object] interface {
	Reconcile(ctx context.Context, client client.Client, parent Parent) (reconcile.Result, error)
	Describe() Descriptor
}

type Descriptor struct {
	Name        string
	Description string
}
