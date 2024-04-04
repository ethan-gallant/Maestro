# Binder Package

The `binder` package provides a convenient way to bind and retrieve values from a `context.Context` using static or
dynamic keys. It offers a type-safe and flexible approach to store and access custom data within a context, which is
particularly useful in the context of Kubernetes controllers and reconcilers.

## Features

- Bind values to a context using static or dynamic keys
- Retrieve values from a context in a type-safe manner
- Unbind values from a context
- Error handling for common scenarios (e.g., value not found, type mismatch)
- Thread-safe and concurrency-friendly

## Usage

The `binder` package offers two main types of bindable objects: `StaticBindable` and `DynamicBindable`.

### StaticBindable

A `StaticBindable` is a generic struct that wraps a value of type `T` and allows binding and retrieval using a static
key. The key is determined by the type of the wrapped value.

Here's an example of how to use a `StaticBindable`:

```go
package main

import (
	"context"
	"github.com/ethan-gallant/maestro/pkg/binder"
	corev1 "k8s.io/api/core/v1"
)

type PodState struct {
	// Define your custom state fields
	Ready bool
}

var podStateBinder = binder.StaticBindable[PodState]{}

func main() {
	ctx := context.Background()
	state := &PodState{Ready: true}

	// Bind the state to the context
	ctx, err := podStateBinder.BindToContext(ctx, state)
	if err != nil {
		// Handle the error
	}

	// Retrieve the state from the context
	retrievedState, err := podStateBinder.FromContext(ctx)
	if err != nil {
		// Handle the error
	}

	// Use the retrieved state
	if retrievedState.Ready {
		// Perform actions based on the state
	}

	// Unbind the state from the context
	ctx = podStateBinder.Unbind(ctx)
}
```

In this example, a `StaticBindable` is used to bind and retrieve a custom `PodState` struct to/from the context.
The `BindToContext` method is used to bind the state to the context, and the `FromContext` method is used to retrieve
the state from the context. The `Unbind` method is used to remove the state from the context when it's no longer needed.

### DynamicBindable

A `DynamicBindable` is similar to a `StaticBindable`, but it allows you to specify a custom key function to generate
unique keys for the bindable value.

Here's an example of how to use a `DynamicBindable`:

```go
package main

import (
	"context"
	"github.com/ethan-gallant/maestro/pkg/binder"
	appsv1 "k8s.io/api/apps/v1"
)

type DeploymentState struct {
	// Define your custom state fields
	Replicas int32
}

func deploymentStateKeyFunc(deployment *appsv1.Deployment) binder.ContextKey {
	return binder.ContextKey("deployment-" + deployment.Name)
}

func main() {
	ctx := context.Background()
	deployment := &appsv1.Deployment{
		// Initialize the Deployment object
	}
	state := &DeploymentState{Replicas: 3}

	// Create a DynamicBindable with the custom key function
	deploymentStateBinder := binder.NewDynamicBindable(func() binder.ContextKey {
		return deploymentStateKeyFunc(deployment)
	})

	// Bind the state to the context
	ctx, err := deploymentStateBinder.BindToContext(ctx, state)
	if err != nil {
		// Handle the error
	}

	// Retrieve the state from the context
	retrievedState, err := deploymentStateBinder.FromContext(ctx)
	if err != nil {
		// Handle the error
	}

	// Use the retrieved state
	if retrievedState.Replicas > 0 {
		// Perform actions based on the state
	}

	// Unbind the state from the context
	ctx = deploymentStateBinder.Unbind(ctx)
}
```

In this example, a `DynamicBindable` is created with a custom key function that generates a unique key based on the name
of a Deployment object. The `BindToContext`, `FromContext`, and `Unbind` methods are used similarly to
the `StaticBindable` example.

### Concurrent Access

The `binder` package is designed to be thread-safe and concurrency-friendly. Multiple goroutines can safely access and
modify the bound values in the context simultaneously. The package ensures that the binding, retrieval, and unbinding
operations are performed atomically to prevent race conditions.

However, it's important to note that the concurrency safety is limited to the operations provided by the `binder`
package itself. If you modify the bound values directly outside of the `binder` package, you need to ensure proper
synchronization mechanisms are in place to avoid race conditions.

## Error Handling

The `binder` package defines several error variables to handle common scenarios:

- `ErrStateNotFound`: Returned when the requested state is not found in the context.
- `ErrStateMismatch`: Returned when the type of the retrieved state doesn't match the expected type.
- `ErrContextExists`: Returned when attempting to bind a value to a context that already contains a value with the same
  key.

Make sure to handle these errors appropriately in your code. You can use the `errors.Is` function from the `errors`
package to check for specific error types:

```go
package main

import "errors"

// ...

retrievedState, err := binder.FromContext(ctx)
if errors.Is(err, binder.ErrStateNotFound) {
// Handle the case when the state is not found
} else if errors.Is(err, binder.ErrStateMismatch) {
// Handle the case when the state type doesn't match
} else if err != nil {
// Handle other errors
}
```

## Integration with Kubernetes Controllers

The `binder` package is particularly useful in the context of Kubernetes controllers and reconcilers. You can use it to
store and retrieve custom state information within the reconciliation context.

For example, let's say you have a custom controller that manages a parent resource (e.g., `Deployment`) and a child
resource (e.g., `Pod`). You can use the `binder` package to store and access the state of the child resources within the
reconciliation context of the parent resource.

```go
package main

import (
	"context"
	"github.com/ethan-gallant/maestro/pkg/binder"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type PodState struct {
	// Define your custom state fields
	Ready bool
}

var podStateBinder = binder.StaticBindable[PodState]{}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	deployment := &appsv1.Deployment{}
	// Retrieve the Deployment object

	// Reconcile the child Pods
	podList := &corev1.PodList{}
	// List the Pods owned by the Deployment

	for _, pod := range podList.Items {
		// Create a new context for each Pod
		podCtx := context.Background()

		// Create and bind the Pod state to the context
		podState := &PodState{Ready: true}
		podCtx, err := podStateBinder.BindToContext(podCtx, podState)
		if err != nil {
			// Handle the error
		}

		// Perform reconciliation logic for the Pod using the podCtx
		// ...

		// Retrieve the Pod state from the context
		retrievedPodState, err := podStateBinder.FromContext(podCtx)
		if err != nil {
			// Handle the error
		}

		// Use the retrieved Pod state
		if retrievedPodState.Ready {
			// Perform actions based on the Pod state
		}
	}

	// Continue with the Deployment reconciliation
	// ...

	return ctrl.Result{}, nil
}
```

In this example, the `binder` package is used to store and retrieve the state of individual Pods within the
reconciliation context of the Deployment controller. This allows for a clean separation of concerns and easy access to
Pod-specific state information during the reconciliation process.

## Best Practices

When using the `binder` package, consider the following best practices:

1. Use meaningful and descriptive names for your custom state types and binder variables to enhance code readability and
   maintainability.

2. Handle errors returned by the `binder` package appropriately. Use the provided error
   variables (`ErrStateNotFound`, `ErrStateMismatch`, `ErrContextExists`) to handle specific error scenarios.

3. Ensure that the bound values are properly synchronized if they are accessed or modified outside of the `binder`
   package. Use appropriate synchronization primitives (e.g., mutexes, channels) to prevent race conditions.

4. Consider using `DynamicBindable` when you need to generate unique keys based on runtime information or when you have
   multiple instances of the same state type.

5. When using the `binder` package in the context of Kubernetes controllers, create a new context for each
   reconciliation iteration to avoid unintended state sharing between different reconciliation cycles.

6. Remember to unbind the state from the context when it's no longer needed to prevent memory leaks and unnecessary
   storage overhead.

## References

For more information on the Kubernetes concepts and resources mentioned in this README, refer to the official Kubernetes
documentation:

- [Kubernetes Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
- [Kubernetes Reconciliation](https://book.kubebuilder.io/cronjob-tutorial/controller-overview.html)
- [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes Pods](https://kubernetes.io/docs/concepts/workloads/pods/)