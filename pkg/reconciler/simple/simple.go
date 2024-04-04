package simple

import (
	"context"
	"fmt"
	"time"

	"github.com/ethan-gallant/maestro/api"
	"github.com/ethan-gallant/maestro/pkg/conductor"
	"github.com/ethan-gallant/maestro/pkg/reconciler"
	"github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler (SimpleReconciler) is a simple reconciler that reconciles a child object for a parent object.
type Reconciler[Parent client.Object, Child client.Object] struct {
	// Details is the descriptor for the reconciler.
	// It should contain the name and description of the reconciler for documentation and debugging purposes.
	Details api.Descriptor // required
	// ReconcileFn is the function that reconciles the Child object.
	// The ReconcileFn accepts a Parent object, and returns the desired state of the Child object, or an error.
	ReconcileFn func(ctx context.Context, parent Parent) (Child, error) // required
	// PredicateFn is a function that returns true if the ReconcileFn should be called.
	// If nil, the ReconcileFn will always be called.
	PredicateFn func(parent Parent) bool // optional
	// NoReference optionally disables setting the owner reference on the child object.
	NoReference bool // optional
	// DryRunType configures the dry-run behavior of the reconciler.
	DryRunType reconciler.DryRunType // optional
	// CompareOpts are the options to use when comparing the child object to the desired state.
	// This helps avoid unnecessary updates when the child object is already in the desired state.
	CompareOpts []cmp.Option // optional
	// ShouldDeleteFn is a function that if returns true, the child object will be deleted.
	// It is called regardless of the PredicateFn function. If no function is provided, the child object will never be deleted.
	ShouldDeleteFn func(Parent) bool // optional
	// ChildKeyFn returns the child object with only a key (name and namespace) set.
	// It must always match the key the ReconcileFn returns. Otherwise, Reconcile calls will fail.
	// All other fields should be empty and will be ignored.
	ChildKeyFn func(Parent) Child // required if ShouldDeleteFn is set
	// PreUpdateFn is a function that is called before the child object is applied.
	// This function is not called for the first creation of the child object.
	PreUpdateFn func(ctx context.Context, parent Parent, previous, child Child) error // optional
}

var _ api.Reconciler[client.Object] = &Reconciler[client.Object, client.Object]{}

// Reconcile method for SimpleReconciler calls the embedded ChildReconciler's Reconcile method and handles the child object.
func (r *Reconciler[Parent, Child]) Reconcile(ctx context.Context, k8sCli client.Client, parent Parent) (reconcile.Result, error) {
	state, err := conductor.FetchState(ctx)
	if err != nil { // With no state / conductor, do a normal reconcile
		return r.doReconcile(ctx, k8sCli, parent)
	}

	result, err := r.doReconcile(ctx, k8sCli, parent)
	if err != nil {
		state.AddCondition(metav1.Condition{
			Type:    fmt.Sprintf("%sError", r.Details.Name),
			Status:  metav1.ConditionTrue,
			Reason:  "ReconcileError",
			Message: err.Error(),
			LastTransitionTime: metav1.Time{
				Time: time.Now(),
			},
		})

		return result, err
	}

	state.AddCondition(metav1.Condition{
		Type:    fmt.Sprintf("%sReconciled", r.Details.Name),
		Status:  conditionFromResult(result),
		Reason:  "Reconciled",
		Message: "Reconciled successfully",
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
	})

	return result, nil
}

// Describe returns the descriptor for the reconciler.
func (r *Reconciler[Parent, Child]) Describe() api.Descriptor {
	return r.Details
}

func conditionFromResult(result reconcile.Result) metav1.ConditionStatus {
	if result.Requeue || result.RequeueAfter > 0 {
		return metav1.ConditionFalse
	}
	return metav1.ConditionTrue
}

func (r *Reconciler[Parent, Child]) doReconcile(ctx context.Context, k8sCli client.Client, parent Parent) (reconcile.Result, error) {
	log := klog.FromContext(ctx).V(1).
		WithValues("parent", client.ObjectKeyFromObject(parent))

	var childKey client.ObjectKey
	if r.ShouldDeleteFn != nil {
		current := r.ChildKeyFn(parent)
		childKey = client.ObjectKeyFromObject(current)
		if err := k8sCli.Get(ctx, client.ObjectKeyFromObject(current), current); err == nil && r.ShouldDeleteFn(parent) {
			if err := k8sCli.Delete(ctx, current); err != nil {
				return reconcile.Result{}, err
			}
			log.Info("deleted child")
			return reconcile.Result{
				Requeue: true,
			}, nil
		} else if err != nil && !apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
	}

	if r.PredicateFn != nil && !r.PredicateFn(parent) {
		return reconcile.Result{}, nil
	}

	desired, err := r.ReconcileFn(ctx, parent)
	if err != nil {
		return reconcile.Result{}, err
	}

	if r.ChildKeyFn != nil {
		// Backfill the name and namespace if not already set by the ReconcileFn
		if desired.GetName() == "" {
			desired.SetName(childKey.Name)
		}
		if desired.GetNamespace() == "" {
			desired.SetNamespace(childKey.Namespace)
		}

		// Error if the there's a mismatch between the key and the object returned by ReconcileFn
		if childKey.Namespace != desired.GetNamespace() || childKey.Name != desired.GetName() {
			return reconcile.Result{}, reconciler.ErrChildKeyMismatch
		}
	}

	key := client.ObjectKeyFromObject(desired)
	log = log.WithValues("child", key.Name, "namespace", key.Namespace, "kind", desired.GetObjectKind().GroupVersionKind().Kind)

	if !r.NoReference {
		if err := controllerutil.SetControllerReference(parent, desired, k8sCli.Scheme()); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Fetch the current object, if not already set from ShouldDeleteFn.
	current := desired.DeepCopyObject().(Child)

	if err := k8sCli.Get(ctx, key, current); err != nil {
		// Allow only not-found errors, any other error is a problem.
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to fetch child")
			return reconcile.Result{}, err
		}

		// Create the object & requeue, it doesn't yet exist.
		if err := k8sCli.Create(ctx, desired); err != nil {
			return reconcile.Result{}, err
		}

		log.Info("created child")
		return reconcile.Result{
			Requeue: true,
		}, nil
	}

	// ResourceVersion should come from the API, so we need to update it.
	// This makes an easier and safer check for changes.
	desired.SetResourceVersion(current.GetResourceVersion())
	desired.SetCreationTimestamp(current.GetCreationTimestamp())
	desired.SetGeneration(current.GetGeneration())
	desired.SetUID(current.GetUID())
	if r.PreUpdateFn != nil {
		if err := r.PreUpdateFn(ctx, parent, current, desired); err != nil {
			return reconcile.Result{}, err
		}
	}

	// We always append the two options IgnoreManagedFields and IgnoreTypeMeta.
	// This avoids unnecessary updates when the child object is already in the desired state.
	compareOpts := append(r.CompareOpts, reconciler.IgnoreManagedFields(), reconciler.IgnoreTypeMeta(), reconciler.IgnoreStatusFields())
	if cmp.Equal(current, desired, compareOpts...) {
		log.Info("no changes", "key", key)
		return reconcile.Result{}, nil
	}

	if r.DryRunType != reconciler.DryRunNone {
		// Dry-run the update to see if it would change anything.
		// We need to copy it due to kubernetes/kubernetes/pull/121167 not being resolved yet.
		// TL;DR, due to the above bug, we need to dry-run both objects (desired and current) then compare them
		desiredCopy := desired.DeepCopyObject().(Child)
		if err := k8sCli.Update(ctx, desiredCopy, client.DryRunAll); err != nil {
			log.Error(err, "unable to dry-run update", "key", key)
			return reconcile.Result{}, err
		}

		// Until kubernetes/kubernetes/pull/121167 is resolved, we need to dry-run as a hack here
		currentHack := current.DeepCopyObject().(Child)
		if err := k8sCli.Update(ctx, currentHack, client.DryRunAll); err != nil {
			log.Error(err, "unable to dry-run update", "key", key)
			return reconcile.Result{}, err
		}

		// When removing after kubernetes/kubernetes/pull/121167 is resolved, swap the currentHack with current
		if cmp.Equal(currentHack, desiredCopy, compareOpts...) {
			// Log the diff, user should update the object returned by ReconcileFn to include the changes.
			// Or to ignore annotations like those added by the deployment controller.
			if r.DryRunType == reconciler.DryRunWarn {
				diff := cmp.Diff(currentHack, desiredCopy, compareOpts...)
				log.Info("no changes after dry-run. Please update CompareOpts or add the API defaults to the object", "diff", diff)
			}

			return reconcile.Result{}, nil
		}
	}

	log.Info("updating child", "key", key)
	// Do an update as it's required.
	if err := k8sCli.Update(ctx, desired); err != nil {
		return reconcile.Result{}, err
	}

	log.Info("updated child", "key", key)
	return reconcile.Result{
		Requeue: true,
	}, nil
}
