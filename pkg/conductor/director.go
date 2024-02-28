package conductor

import (
	"context"
	"github.com/ethan-gallant/maestro/api"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Director[Parent client.Object] struct {
	client      client.Client
	ctx         context.Context
	parent      Parent
	log         klog.Logger
	reconcilers []api.Reconciler[Parent]
}

func (d *Director[Parent]) Register(reconciler api.Reconciler[Parent]) api.Conductor[Parent] {
	d.reconcilers = append(d.reconcilers, reconciler)
	return d
}

func (d *Director[Parent]) Conduct(parent Parent) (reconcile.Result, error) {
	d.parent = parent
	for _, reconciler := range d.reconcilers {
		result, err := d.Reconcile(reconciler)
		if err != nil {
			return result, err
		}
	}
	return reconcile.Result{}, nil
}

var _ api.Conductor[client.Object] = &Director[client.Object]{}

// Reconcile takes a reconciler and invokes its Reconcile method, providing the necessary dependencies.
func (d *Director[Parent]) Reconcile(
	// Any type for the second arg
	reconciler api.Reconciler[Parent],
) (reconcile.Result, error) {
	return reconciler.Reconcile(d.ctx, d.client, d.parent)
}
