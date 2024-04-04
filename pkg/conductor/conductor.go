package conductor

import (
	"context"

	"github.com/ethan-gallant/maestro/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Conductor[Parent client.Object] struct {
	client            client.Client
	ctx               context.Context
	parent            Parent
	log               klog.Logger
	reconcilers       []api.Reconciler[Parent]
	conditionsHandler StatusConditionHandler
}

type StatusConditionHandler func(ctx context.Context, client client.Client, parent client.Object, conditions []metav1.Condition) error

var _ api.Conductor[client.Object] = &Conductor[client.Object]{}

func (d *Conductor[Parent]) Register(reconciler api.Reconciler[Parent]) api.Conductor[Parent] {
	d.reconcilers = append(d.reconcilers, reconciler)
	return d
}

func (d *Conductor[Parent]) Conduct(ctx context.Context, parent Parent) (reconcile.Result, error) {
	state := &State{
		Conditions: []metav1.Condition{},
	}
	if _, err := BindState(ctx, state); err != nil {
		return reconcile.Result{}, err
	}

	d.parent = parent
	for _, reconciler := range d.reconcilers {
		if result, err := d.Reconcile(state.ctx, reconciler); shouldReturn(result, err) {
			return result, err
		}
	}

	if d.conditionsHandler != nil {
		if err := d.conditionsHandler(state.ctx, d.client, parent, state.Conditions); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// Reconcile takes a single reconciler and invokes its Reconcile method, providing the necessary dependencies.
func (d *Conductor[Parent]) Reconcile(
	ctx context.Context,
	reconciler api.Reconciler[Parent],
) (reconcile.Result, error) {
	return reconciler.Reconcile(ctx, d.client, d.parent)
}

func shouldReturn(result reconcile.Result, err error) bool {
	return err != nil || result.Requeue || result.RequeueAfter > 0
}
