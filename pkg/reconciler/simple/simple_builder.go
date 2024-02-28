package simple

import (
	"context"
	"github.com/ethan-gallant/maestro/api"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconcileFn[Parent client.Object, Child client.Object] func(ctx context.Context, parent Parent) (Child, error)

// Builder is a builder for the Reconciler.
type Builder[Parent client.Object, Child client.Object] struct {
	reconciler Reconciler[Parent, Child]
}

// FromReconcileFunc returns a new instance of Builder for the ReconcileFn
func FromReconcileFunc[Parent client.Object, Child client.Object](fn ReconcileFn[Parent, Child]) *Builder[Parent, Child] {
	return &Builder[Parent, Child]{
		reconciler: Reconciler[Parent, Child]{
			ReconcileFn: fn,
		},
	}
}

// WithPredicate sets the Predicate field.
func (b *Builder[Parent, Child]) WithPredicate(predicate func(parent Parent) bool) *Builder[Parent, Child] {
	b.reconciler.Predicate = predicate
	return b
}

// WithNoReference sets the NoReference field.
func (b *Builder[Parent, Child]) WithNoReference(noReference bool) *Builder[Parent, Child] {
	b.reconciler.NoReference = noReference
	return b
}

// WithDryRunType configures the dry-run behavior of the reconciler.
func (b *Builder[Parent, Child]) WithDryRunType(dryRunType DryRunType) *Builder[Parent, Child] {
	b.reconciler.DryRunType = dryRunType
	return b
}

// AddCompareOpt adds a comparator option to the reconciler
func (b *Builder[Parent, Child]) AddCompareOpt(compareOpts []cmp.Option) *Builder[Parent, Child] {
	b.reconciler.CompareOpts = append(b.reconciler.CompareOpts, compareOpts...)
	return b
}

// WithDetails sets the Details field.
func (b *Builder[Parent, Child]) WithDetails(details api.Descriptor) *Builder[Parent, Child] {
	b.reconciler.Details = details
	return b
}

// Build returns the constructed Reconciler.
func (b *Builder[Parent, Child]) Build() *Reconciler[Parent, Child] {
	return &b.reconciler
}
