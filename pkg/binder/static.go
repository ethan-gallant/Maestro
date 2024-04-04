package binder

import "context"

// StaticBindable is a generic struct that wraps a value of type T and allows Static key assignment
type StaticBindable[T any] struct {
	// Key should really never be set, it is used as a blank object to fetch things from the context.
	key *T
}

// BindToContext binds the StaticBindable value to the provided context using the stored key
func (b *StaticBindable[T]) BindToContext(ctx context.Context, value *T) (context.Context, error) {
	if ctx.Value(b.key) != nil {
		return nil, ErrContextExists
	}

	return context.WithValue(ctx, b.key, value), nil
}

// Unbind removes the StaticBindable value from the provided context using the stored key
func (b *StaticBindable[T]) Unbind(ctx context.Context) context.Context {
	return context.WithValue(ctx, b.key, nil)
}

// FromContext retrieves the StaticBindable value from the provided context using the stored key
func (b *StaticBindable[T]) FromContext(ctx context.Context) (*T, error) {
	if ctx.Value(b.key) == nil {
		return nil, ErrStateNotFound
	}
	if value, ok := ctx.Value(b.key).(*T); ok {
		return value, nil
	}
	return nil, ErrStateMismatch
}
