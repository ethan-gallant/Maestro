package conductor

import (
	"context"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder[Parent client.Object] struct {
	client client.Client
	ctx    context.Context
	parent Parent
	log    klog.Logger
}

func ForParent[Parent client.Object](parent Parent) *Builder[Parent] {
	return &Builder[Parent]{
		ctx:    context.Background(),
		parent: parent,
	}
}

func (b *Builder[Parent]) WithClient(c client.Client) *Builder[Parent] {
	b.client = c
	return b
}

func (b *Builder[Parent]) WithContext(ctx context.Context) *Builder[Parent] {
	b.ctx = ctx
	return b
}

func (b *Builder[Parent]) WithLogger(l klog.Logger) *Builder[Parent] {
	b.log = l
	return b
}

func (b *Builder[Parent]) Build() *Director[Parent] {
	return &Director[Parent]{
		client: b.client,
		ctx:    b.ctx,
		parent: b.parent,
		log:    b.log,
	}
}
