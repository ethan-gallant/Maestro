package api

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Conductor is an interface to simplify registration, execution and management of reconcilers
type Conductor[Parent client.Object] interface {
	// Register takes a reconciler and registers it with the director
	Register(reconciler Reconciler[Parent]) Conductor[Parent]
	// Conduct takes the parent object and invokes the reconcilers registered with the director
	Conduct(parent Parent) (reconcile.Result, error)
}
