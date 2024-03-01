package conductor

import (
	"context"
	"github.com/ethan-gallant/maestro/api"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Conductor[Parent client.Object] struct {
	client      client.Client
	ctx         context.Context
	parent      Parent
	log         klog.Logger
	reconcilers []api.Reconciler[Parent]
}

var _ api.Conductor[client.Object] = &Conductor[client.Object]{}

func (d *Conductor[Parent]) Register(reconciler api.Reconciler[Parent]) api.Conductor[Parent] {
	d.reconcilers = append(d.reconcilers, reconciler)
	return d
}

func (d *Conductor[Parent]) Conduct(parent Parent) (reconcile.Result, error) {
	d.parent = parent
	for _, reconciler := range d.reconcilers {
		if result, err := d.Reconcile(reconciler); shouldReturn(result, err) {
			return result, err
		}
	}
	return reconcile.Result{}, nil
}

// Reconcile takes a single reconciler and invokes its Reconcile method, providing the necessary dependencies.
func (d *Conductor[Parent]) Reconcile(
	// Any type for the second arg
	reconciler api.Reconciler[Parent],
) (reconcile.Result, error) {
	return reconciler.Reconcile(d.ctx, d.client, d.parent)
}

func shouldReturn(result reconcile.Result, err error) bool {
	return err != nil || result.Requeue || result.RequeueAfter > 0
}
