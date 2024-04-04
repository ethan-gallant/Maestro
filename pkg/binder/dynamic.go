package binder

import "context"

// DynamicBindable is a generic struct that wraps a value of type T and allows dynamic key assignment
type DynamicBindable[T any] struct {
	// keyFunc is a function that generates a unique key for the bindable value
	keyFunc func() ContextKey
}

// NewDynamicBindable creates a new DynamicBindable instance with the provided key function
func NewDynamicBindable[T any](keyFunc func() ContextKey) *DynamicBindable[T] {
	return &DynamicBindable[T]{
		keyFunc: keyFunc,
	}
}

// BindToContext binds the DynamicBindable value to the provided context using the dynamic key
func (b *DynamicBindable[T]) BindToContext(ctx context.Context, value *T) (context.Context, error) {
	key := b.keyFunc()
	if ctx.Value(key) != nil {
		return nil, ErrContextExists
	}

	return context.WithValue(ctx, key, value), nil
}

// Unbind removes the DynamicBindable value from the provided context using the dynamic key
func (b *DynamicBindable[T]) Unbind(ctx context.Context) context.Context {
	key := b.keyFunc()
	return context.WithValue(ctx, key, nil)
}

// FromContext retrieves the DynamicBindable value from the provided context using the dynamic key
func (b *DynamicBindable[T]) FromContext(ctx context.Context) (*T, error) {
	key := b.keyFunc()
	if ctx.Value(key) == nil {
		return nil, ErrStateNotFound
	}
	if value, ok := ctx.Value(key).(*T); ok {
		return value, nil
	}
	return nil, ErrStateMismatch
}
