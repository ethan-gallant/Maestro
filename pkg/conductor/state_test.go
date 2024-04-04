package conductor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddCondition(t *testing.T) {
	state := &State{}

	condition1 := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}
	condition2 := metav1.Condition{
		Type:   "Synced",
		Status: metav1.ConditionFalse,
	}

	state.AddCondition(condition1)
	state.AddCondition(condition2)

	assert.Len(t, state.Conditions, 2)
	assert.Contains(t, state.Conditions, condition1)
	assert.Contains(t, state.Conditions, condition2)
}

func TestBindState(t *testing.T) {
	ctx := context.Background()
	state := &State{}

	ctx, err := BindState(ctx, state)
	require.NoError(t, err)

	fetchedState, err := FetchState(ctx)
	require.NoError(t, err)
	assert.Same(t, state, fetchedState)
}

func TestClearState(t *testing.T) {
	ctx := context.Background()
	state := &State{}

	ctx, err := BindState(ctx, state)
	require.NoError(t, err)

	ctx = ClearState(ctx)

	fetchedState, err := FetchState(ctx)
	assert.Error(t, err)
	assert.Nil(t, fetchedState)
}

func TestFetchState(t *testing.T) {
	ctx := context.Background()
	state := &State{}

	_, err := FetchState(ctx)
	assert.Error(t, err)

	ctx, err = BindState(ctx, state)
	require.NoError(t, err)

	fetchedState, err := FetchState(ctx)
	require.NoError(t, err)
	assert.Same(t, state, fetchedState)
}

func TestConcurrentAddCondition(t *testing.T) {
	state := &State{}

	condition1 := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}
	condition2 := metav1.Condition{
		Type:   "Synced",
		Status: metav1.ConditionFalse,
	}

	var counter int
	for i := 0; i < 100; i++ {
		go func() {
			state.AddCondition(condition1)
			state.AddCondition(condition2)
			counter++
		}()
	}

	assert.Eventually(t, func() bool {
		return counter == 100
	}, 5*time.Second, 100*time.Millisecond)

	assert.Len(t, state.Conditions, 200)
}
