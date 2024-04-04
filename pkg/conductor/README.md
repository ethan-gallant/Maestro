# Conductor

The Conductor package provides a framework for orchestrating the reconciliation process of Kubernetes custom resources (
CRs) using a parent-child relationship. It allows you to define a set of reconcilers that operate on a parent object and
its associated child objects, ensuring a structured and organized approach to resource management.

## Features

- Supports generic parent objects that implement the `client.Object` interface (
  see [Kubernetes API Machinery](https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L93))
- Allows registration of multiple reconcilers for a parent object
- Provides a convenient builder pattern for constructing a conductor instance
- Supports custom status condition handling for the parent object (
  see [Status Condition Handling](#status-condition-handling))
- Utilizes a state management mechanism to share data across reconcilers (
  see [Custom State Management](#custom-state-management))
- Enables custom state management using the `pkg/binder` package (
  see [pkg/binder](https://github.com/ethan-gallant/maestro/tree/master/pkg/binder))

## Usage

To use the Conductor package, follow these steps:

1. Define your parent and child objects as Kubernetes custom resources. For example, let's consider a `Deployment` as
   the parent object (
   see [Kubernetes Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)) and a `Pod` as
   the child object (see [Kubernetes Pod](https://kubernetes.io/docs/concepts/workloads/pods/)).

2. Implement the necessary reconcilers for your parent object. Each reconciler should implement the `api.Reconciler`
   interface (see [api package](https://github.com/ethan-gallant/maestro/tree/master/api)), which defines a `Reconcile`
   method. The reconciler is responsible for performing specific operations on the parent and child objects.

3. Create a new conductor instance using the `ForParent` function, specifying the parent object type. For example:

   ```go
   conductor := conductor.ForParent(&appsv1.Deployment{})
   ```

4. Configure the conductor using the available builder methods:
    - `WithClient`: Set the Kubernetes client for interacting with the API server (
      see [controller-runtime client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client)).
    - `WithContext`: Set the context for the reconciliation process (
      see [Go context package](https://pkg.go.dev/context)).
    - `WithLogger`: Set the logger for logging purposes (see [klog package](https://pkg.go.dev/k8s.io/klog/v2)).
    - `WithStatusConditionsHandler`: Set a custom handler for updating the status conditions of the parent object (
      see [Status Condition Handling](#status-condition-handling)).

5. Register your reconcilers with the conductor using the `Register` method. For example:

   ```go
   conductor.Register(&PodReconciler{})
   ```

6. Invoke the `Conduct` method on the conductor, passing the parent object. The conductor will execute the registered
   reconcilers in the order they were registered.

   ```go
   result, err := conductor.Conduct(ctx, deployment)
   ```

7. Handle the reconciliation result and error as needed (
   see [controller-runtime reconcile package](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile)).

## Status Condition Handling

The Conductor package provides a mechanism for handling and updating the status conditions of the parent object. Status
conditions are a way to represent the current state of the object and provide meaningful information about its condition
to users or other components of the system.

### Adding Custom Status Conditions

To add custom status conditions to the parent object, you can utilize the `State` object provided by the Conductor
package. The `State` object is stored in the context and can be accessed within your reconcilers.

Here's an example of how to add a custom status condition to the `State`:

```go
package main

func (r *MyReconciler) Reconcile(
	ctx context.Context,
	client client.Client,
	parent *MyParent
) (reconcile.Result, error) {
	// ...

	state, _ := conductor.FetchState(ctx)
	state.AddCondition(metav1.Condition{
		Type:    "MyCondition",
		Status:  metav1.ConditionTrue,
		Reason:  "ConditionReason",
		Message: "Custom condition message",
	})

	// ...
}
```

In this example, we fetch the `State` object from the context using `conductor.FetchState(ctx)`. We then use
the `AddCondition` method to add a custom condition to the state. The `metav1.Condition` struct represents a single
condition and includes fields such as `Type`, `Status`, `Reason`, and `Message` (
see [Kubernetes API Machinery Condition](https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go#L1423)).

You can add multiple conditions to the state by calling `AddCondition` multiple times with different condition objects.

### Registering a Status Condition Update Function

To ensure that the status conditions added to the `State` are actually updated on the parent object, you need to
register a status condition update function when building the conductor.

Here's an example of how to register a status condition update function:

```go
package main

func main() {
	// ...

	conductor := conductor.ForParent(parent).
		// ...
		WithStatusConditionsHandler(func(
			ctx context.Context,
			client client.Client,
			parent *MyParent,
			conditions []metav1.Condition
		) error {
			// Update the status conditions of the parent object
			parent.Status.Conditions = conditions
			return client.Status().Update(ctx, parent)
		}).
		Build()

	// ...
}
```

In this example, we use the `WithStatusConditionsHandler` method when building the conductor to register a status
condition update function. The function takes the context, client, parent object, and a slice of `metav1.Condition`
objects as parameters.

Inside the function, you can update the status conditions of the parent object based on the provided conditions. In this
case, we directly assign the conditions to the `parent.Status.Conditions` field. Finally, we use
the `client.Status().Update(ctx, parent)` method to update the status of the parent object in the Kubernetes API
server (
see [controller-runtime client status package](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#StatusWriter)).

It's important to note that if you don't register a status condition update function, the conditions added to
the `State` will be discarded and not persisted on the parent object.

### Example Usage

Here's a complete example demonstrating the usage of status condition handling in the Conductor package:

```go
package main

import (
	// ...
	"github.com/ethan-gallant/maestro/conductor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// ...
)

type MyReconciler struct{}

func (r *MyReconciler) Reconcile(
	ctx context.Context,
	client client.Client,
	parent *MyParent
) (reconcile.Result, error) {
	// ...

	state, _ := conductor.FetchState(ctx)
	state.AddCondition(metav1.Condition{
		Type:    "MyCondition",
		Status:  metav1.ConditionTrue,
		Reason:  "ConditionReason",
		Message: "Custom condition message",
	})

	// ...
}

func main() {
	parent := &MyParent{
		// ...
	}

	conductor := conductor.ForParent(parent).
		WithClient(client).
		WithContext(ctx).
		WithLogger(logger).
		WithStatusConditionsHandler(func(
			ctx context.Context,
			client client.Client,
			parent *MyParent,
			conditions []metav1.Condition
		) error {
			parent.Status.Conditions = conditions
			return client.Status().Update(ctx, parent)
		}).
		Build()

	conductor.Register(&MyReconciler{})

	result, err := conductor.Conduct(ctx, parent)
	// Handle the result and error
	// ...
}
```

In this example, the `MyReconciler` adds a custom status condition to the `State` during reconciliation. The registered
status condition update function in the conductor builder updates the status conditions of the parent object based on
the conditions in the `State`.

By leveraging the status condition handling mechanism provided by the Conductor package, you can effectively manage and
communicate the state of your parent objects, providing valuable information to users and other components of your
system.

## Custom State Management

In addition to the built-in state management provided by the Conductor package, you can also define and utilize custom
state objects within your reconcilers. This allows you to share data and state across multiple reconcilers, enabling
more complex reconciliation scenarios.

To define a custom state object, you can create a new type that implements the necessary methods for binding and
retrieving the state from the context. The `pkg/binder` package provides a `StaticBindable` type that you can use as a
starting point (see [pkg/binder](https://github.com/ethan-gallant/maestro/tree/master/pkg/binder)).

Here's an example of how you can define a custom state object and use it within your reconcilers:

1. Define your custom state object:

   ```go
   type MyCustomState struct {
       Data string
   }
   
   var myCustomStateBinder = binder.StaticBindable[MyCustomState]{}
   ```

   In this example, we define a `MyCustomState` struct that holds some custom data. We also create a new instance
   of `StaticBindable` specific to our custom state type.

2. Inside a reconciler, bind the custom state to the context and update the state's context:

   ```go
   func (r *MyReconciler) Reconcile(ctx context.Context, client client.Client, parent *MyParent) (reconcile.Result, error) {
       // ...
   
       customState := &MyCustomState{
           Data: "some data",
       }
   
       ctx, err := myCustomStateBinder.BindToContext(ctx, customState)
       if err != nil {
           return reconcile.Result{}, err
       }
   
       state, _ := conductor.FetchState(ctx)
       state.UpdateContext(ctx)
   
       // ...
   }
   ```

   In this example, we create an instance of our custom state object and bind it to the context using
   the `BindToContext` method of our `myCustomStateBinder`. We then fetch the conductor's state
   using `conductor.FetchState(ctx)` and update the state's context with the new context that includes the custom state.

3. In a later reconciler, retrieve the custom state from the context:

   ```go
   func (r *AnotherReconciler) Reconcile(ctx context.Context, client client.Client, parent *MyParent) (reconcile.Result, error) {
       // ...
   
       customState, err := myCustomStateBinder.FromContext(ctx)
       if err != nil {
           return reconcile.Result{}, err
       }
   
       // Use the custom state data
       fmt.Println(customState.Data)
   
       // ...
   }
   ```

   In this example, we retrieve the custom state object from the context using the `FromContext` method of
   our `myCustomStateBinder`. We can then access the data stored in the custom state object and use it within the
   reconciler.

By defining custom state objects and utilizing the `StaticBindable` type from the `pkg/binder` package, you can easily
share data and state across multiple reconcilers. This allows for more complex reconciliation scenarios and enables
better coordination between reconcilers.

Remember to handle errors appropriately when binding and retrieving custom state objects from the context, as failure to
do so may lead to unexpected behavior or panics.

## Example

Here's an example of how to use the Conductor package to reconcile a Deployment and its associated Pods:

```go
package main

import (
	"context"
	"github.com/ethan-gallant/maestro/api"
	"github.com/ethan-gallant/maestro/pkg/conductor"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PodReconciler struct{}

func (r *PodReconciler) Reconcile(
	ctx context.Context,
	client client.Client,
	deployment *appsv1.Deployment
) (reconcile.Result, error) {
	// Reconcile logic for Pods associated with the Deployment
	// ...

	state, _ := conductor.FetchState(ctx)
	state.AddCondition(metav1.Condition{
		Type:    "PodReconciled",
		Status:  metav1.ConditionTrue,
		Reason:  "PodsReconciled",
		Message: "All Pods have been reconciled",
	})

	return reconcile.Result{}, nil
}

func main() {
	deployment := &appsv1.Deployment{
		// ...
	}

	conductor := conductor.ForParent(deployment).
		WithClient(client).
		WithContext(ctx).
		WithLogger(logger).
		WithStatusConditionsHandler(func(
			ctx context.Context,
			client client.Client,
			deployment *appsv1.Deployment,
			conditions []metav1.Condition
		) error {
			// Update the status conditions of the Deployment
			// ...
			return nil
		}).
		Build()

	conductor.Register(&PodReconciler{})

	result, err := conductor.Conduct(ctx, deployment)
	// Handle the result and error
	// ...
}
```

In this example, the `PodReconciler` is registered with the conductor to reconcile the Pods associated with a
Deployment. The reconciler adds a status condition to indicate that the Pods have been reconciled. The custom status
condition handler is responsible for updating the status conditions of the Deployment based on the conditions provided
by the reconcilers.

For more information on Kubernetes concepts and resources mentioned in this README, refer to the official Kubernetes
documentation:

- [Kubernetes API Machinery](https://kubernetes.io/docs/concepts/overview/kubernetes-api/)
- [Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes Pods](https://kubernetes.io/docs/concepts/workloads/pods/)