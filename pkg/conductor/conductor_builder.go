package conductor

import (
	"context"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder[Parent client.Object] struct {
	conductor *Conductor[Parent]
}

func ForParent[Parent client.Object](parent Parent) *Builder[Parent] {
	return &Builder[Parent]{
		conductor: &Conductor[Parent]{
			parent: parent,
			ctx:    context.Background(),
		},
	}
}

func (b *Builder[Parent]) WithClient(c client.Client) *Builder[Parent] {
	b.conductor.client = c
	return b
}

func (b *Builder[Parent]) WithContext(ctx context.Context) *Builder[Parent] {
	b.conductor.ctx = ctx
	return b
}

func (b *Builder[Parent]) WithLogger(l klog.Logger) *Builder[Parent] {
	b.conductor.log = l
	return b
}

func (b *Builder[Parent]) WithStatusConditionsHandler(handler StatusConditionHandler) *Builder[Parent] {
	b.conductor.conditionsHandler = handler
	return b
}

func (b *Builder[Parent]) Build() *Conductor[Parent] {
	// Return an identical copy of the conductor (to prevent mutation)
	return &Conductor[Parent]{
		client:            b.conductor.client,
		ctx:               b.conductor.ctx,
		parent:            b.conductor.parent,
		log:               b.conductor.log,
		reconcilers:       b.conductor.reconcilers,
		conditionsHandler: b.conductor.conditionsHandler,
	}
}
