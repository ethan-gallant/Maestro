package simple

import (
	"context"
	"github.com/ethan-gallant/maestro/api"
	"github.com/ethan-gallant/maestro/pkg/reconciler"
	"github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	klog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler (SimpleReconciler) is a simple reconciler that reconciles a child object for a parent object.
type Reconciler[Parent client.Object, Child client.Object] struct {
	// ReconcileFn is the function that reconciles the child object.
	// It should return the child object and an error if it fails for some reason.
	ReconcileFn func(ctx context.Context, parent Parent) (Child, error)
	// Predicate is a function that returns true if the child should be reconciled.
	Predicate func(parent Parent) bool
	// NoReference optionally disables setting the owner reference on the child object.
	NoReference bool
	// DryRunType configures the dry-run behavior of the reconciler.
	DryRunType reconciler.DryRunType
	// CompareOpts are the options to use when comparing the child object to the desired state.
	// This helps avoid unnecessary updates when the child object is already in the desired state.
	CompareOpts []cmp.Option
	// Details is the descriptor for the reconciler.
	// It should contain the name and description of the reconciler for documentation and debugging purposes.
	Details api.Descriptor
}

var _ api.Reconciler[client.Object] = &Reconciler[client.Object, client.Object]{}

// Reconcile method for SimpleReconciler calls the embedded ChildReconciler's Reconcile method and handles the child object.
func (r *Reconciler[Parent, Child]) Reconcile(ctx context.Context, k8sCli client.Client, parent Parent) (reconcile.Result, error) {
	if r.Predicate != nil && !r.Predicate(parent) {
		return reconcile.Result{}, nil
	}

	log := klog.FromContext(ctx).V(1).
		WithValues("parent", parent.GetName(), "namespace", parent.GetNamespace())

	desired, err := r.ReconcileFn(ctx, parent)
	if err != nil {
		return reconcile.Result{}, err
	}

	key := client.ObjectKeyFromObject(desired)
	log = log.WithValues("child", key.Name, "namespace", key.Namespace, "kind", desired.GetObjectKind().GroupVersionKind().Kind)

	if !r.NoReference {
		if err := controllerutil.SetControllerReference(parent, desired, k8sCli.Scheme()); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Use a copy of the existin object for type-inference, properties/fields are replaced by Get() below.
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
	desired.SetUID(current.GetUID())

	// We always append the two options IgnoreManagedFields and IgnoreTypeMeta.
	// This avoids unnecessary updates when the child object is already in the desired state.
	compareOpts := append(r.CompareOpts, reconciler.IgnoreManagedFields, reconciler.IgnoreTypeMeta)
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
			return reconcile.Result{}, nil
		} else {
			// Log the diff, user should update the object returned by ReconcileFn to include the changes.
			// Or to ignore annotations like the deployment controller does.
			diff := cmp.Diff(currentHack, desiredCopy, compareOpts...)
			if r.DryRunType == reconciler.DryRunWarn {
				log.Info("no changes after dry-run. Please update CompareOpts or add the API defaults to the object", "diff", diff)
			}
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

// Describe returns the descriptor for the reconciler.
func (r *Reconciler[Parent, Child]) Describe() api.Descriptor {
	return r.Details
}
