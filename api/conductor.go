package api

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Conductor is an interface to simplify registration, execution and management of reconcilers
type Conductor[Parent client.Object] interface {
	// Register takes a reconciler and registers it with the director
	Register(reconciler Reconciler[Parent]) Conductor[Parent]
	// Conduct takes the parent object and invokes the reconcilers registered with the director
	Conduct(ctx context.Context, parent Parent) (reconcile.Result, error)
	// Reconcile takes a single reconciler and invokes its Reconcile method, providing the necessary dependencies.
	Reconcile(ctx context.Context, reconciler Reconciler[Parent]) (reconcile.Result, error)
}
