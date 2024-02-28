package simple

import (
	"context"
	"github.com/ethan-gallant/maestro/api"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DryRunType string

const (
	// DryRunWarn will attempt a dry-run on a mismatch and log a warning if the object is identical after the dry-run (default)
	DryRunWarn DryRunType = "warn"
	// DryRunSilent will perform a dry-run and not log anything if the object is identical after the dry-run
	DryRunSilent DryRunType = "silent"
	// DryRunNone will not perform a dry-run, and will always update the object if it is different
	DryRunNone DryRunType = "none"
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
	DryRunType DryRunType
	// CompareOpts are the options to use when comparing the child object to the desired state.
	// This helps avoid unnecessary updates when the child object is already in the desired state.
	CompareOpts []cmp.Option
	// Details is the descriptor for the reconciler.
	// It should contain the name and description of the reconciler for documentation and debugging purposes.
	Details api.Descriptor
}

var _ api.Reconciler[client.Object] = &Reconciler[client.Object, client.Object]{}

var IgnoreManagedFields = cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ManagedFields")

// Reconcile method for SimpleReconciler calls the embedded ChildReconciler's Reconcile method and handles the child object.
func (r *Reconciler[Parent, Child]) Reconcile(ctx context.Context, k8sCli client.Client, parent Parent) (reconcile.Result, error) {
	log := klog.FromContext(ctx)
	if r.Predicate != nil && !r.Predicate(parent) {
		return reconcile.Result{}, nil
	}

	debugLog := log.WithValues("parent", parent.GetName(), "namespace", parent.GetNamespace()).V(1)

	obj, err := r.ReconcileFn(ctx, parent)
	if err != nil {
		return reconcile.Result{}, err
	}

	key := client.ObjectKeyFromObject(obj)
	debugLog = debugLog.WithValues("child", key.Name, "namespace", key.Namespace, "kind", obj.GetObjectKind().GroupVersionKind().Kind)

	if !r.NoReference {
		if err := controllerutil.SetControllerReference(parent, obj, k8sCli.Scheme()); err != nil {
			return reconcile.Result{}, err
		}
	}

	desired := obj.DeepCopyObject().(Child)

	if err := k8sCli.Get(ctx, key, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			debugLog.Error(err, "unable to fetch child")
			return reconcile.Result{}, err
		}

		if err := k8sCli.Create(ctx, obj); err != nil {
			return reconcile.Result{}, err
		}

		debugLog.Info("created child")
		return reconcile.Result{
			Requeue: true,
		}, nil
	}

	// ResourceVersion should come from the API, so we need to update it.
	desired.SetResourceVersion(obj.GetResourceVersion())
	desired.SetCreationTimestamp(obj.GetCreationTimestamp())
	desired.SetUID(obj.GetUID())

	// This allows us to set the TypeMeta field on the object.
	kind, err := apiutil.GVKForObject(obj, k8sCli.Scheme())
	if err != nil {
		return reconcile.Result{}, err
	}
	desired.GetObjectKind().SetGroupVersionKind(kind)

	compareOpts := append(r.CompareOpts, IgnoreManagedFields)
	if cmp.Equal(obj, desired, compareOpts...) {
		debugLog.Info("no changes", "key", key)
		return reconcile.Result{}, nil
	}

	if r.DryRunType != DryRunNone {
		// Dry-run the update to see if it would change anything.
		// We need to copy it due to kubernetes/kubernetes/pull/121167 not being resolved yet.
		// TL;DR, due to the above bug, we need to dry-run both objects (desired and current) then compare them
		desiredCopy := desired.DeepCopyObject().(client.Object)
		if err := k8sCli.Update(ctx, desiredCopy, client.DryRunAll); err != nil {
			debugLog.Error(err, "unable to dry-run update", "key", key)
			return reconcile.Result{}, err
		}

		// Until kubernetes/kubernetes/pull/121167 is resolved, we need to dry-run as a hack here
		if err := k8sCli.Update(ctx, obj, client.DryRunAll); err != nil {
			debugLog.Error(err, "unable to dry-run update", "key", key)
			return reconcile.Result{}, err
		}

		if cmp.Equal(obj, desiredCopy, compareOpts...) {
			// Log the diff, user should update the object returned by ReconcileFn to include the changes.
			// Or to ignore annotations like the deployment controller does.
			diff := cmp.Diff(obj, desiredCopy, compareOpts...)
			if r.DryRunType == DryRunWarn {
				debugLog.Info("no changes after dry-run. Please update CompareOpts or add the API defaults to the object", "diff", diff)
			}
			return reconcile.Result{}, nil
		}
	}

	debugLog.Info("updating child", "key", key)
	// Do an update as it's required.
	if err := k8sCli.Update(ctx, desired); err != nil {
		return reconcile.Result{}, err
	}

	debugLog.Info("updated child", "key", key)
	return reconcile.Result{
		Requeue: true,
	}, nil
}

// Describe returns the descriptor for the reconciler.
func (r *Reconciler[Parent, Child]) Describe() api.Descriptor {
	return r.Details
}
