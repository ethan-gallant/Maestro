package conductor

import (
	"context"
	"sync"

	"github.com/ethan-gallant/maestro/pkg/binder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var contextBinder = binder.StaticBindable[State]{}

type State struct {
	Conditions []metav1.Condition
	sync.Mutex
	ctx context.Context
}

func (s *State) AddCondition(condition metav1.Condition) {
	s.Lock()
	defer s.Unlock()
	s.Conditions = append(s.Conditions, condition)
}

func (s *State) UpdateContext(ctx context.Context) {
	s.Lock()
	defer s.Unlock()
	s.ctx = ctx
}

func BindState(ctx context.Context, state *State) (context.Context, error) {
	ctx, err := contextBinder.BindToContext(ctx, state)
	if err != nil {
		return nil, err
	}
	state.UpdateContext(ctx) // back-reference the context in the state
	return ctx, nil
}

func ClearState(ctx context.Context) context.Context {
	return contextBinder.Unbind(ctx)
}

func FetchState(ctx context.Context) (*State, error) {
	return contextBinder.FromContext(ctx)
}
