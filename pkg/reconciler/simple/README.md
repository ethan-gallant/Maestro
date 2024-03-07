# Simple Reconciler Package

The Simple Reconciler package provides an opinionated and streamlined approach to building reconcilers for Kubernetes
custom resources. It offers a set of powerful features and utilities to simplify the reconciliation process and enhance
the overall development experience.

## Features

- Reconciles a child object for a parent object using a simple and intuitive API
- Supports generic parent and child objects that implement
  the [`client.Object`](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#Object) interface
- Provides a fluent builder pattern for constructing reconcilers with ease
- Offers customizable predicate functions to control reconciliation behavior
- Supports optional owner references for child objects
- Enables dry-run mode for avoiding unnecessary requeues and optimizing performance
- Allows customization of object comparison options to avoid unnecessary updates
- Supports deletion of child objects based on custom conditions
- Integrates with the [Conductor package](https://github.com/ethan-gallant/maestro/tree/master/pkg/conductor) for advanced
  status condition handling

## Usage

To use the Simple Reconciler package, follow these steps:

1. Define your parent and child objects
   as [Kubernetes custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/),
   ensuring they implement the [`client.Object`](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#Object)
   interface.

2. Create a reconcile function that accepts a parent object and returns the desired state of the child object. The
   function should have the following signature:
   ```go
   func(ctx context.Context, parent Parent) (Child, error)
   ```

3. Use the `FromReconcileFunc` function to create a new reconciler builder, passing in your reconcile function:
   ```go
   builder := simple.FromReconcileFunc(func(ctx context.Context, parent *myapi.MyParent) (*myapi.MyChild, error) {
       // Reconcile logic goes here
   })
   ```

4. Customize the reconciler behavior using the available builder methods:
    - `WithPredicateFn`: Set a predicate function to control when the reconcile function should be called.
    - `WithNoReference`: Disable setting the owner reference on the child object.
    - `WithDryRunType`: Configure the dry-run behavior of the reconciler for avoiding unnecessary requeues and
      optimizing performance.
    - `AddCompareOpt`: Add custom comparison options to avoid unnecessary updates.
    - `WithDetails`: Set the reconciler details, including name and description, for documentation and debugging
      purposes.
    - `WithShouldDeleteFn`: Specify a function to determine when the child object should be deleted.
    - `WithChildKeyFn`: Set a function to return the child object with only a key (name and namespace) set.

5. Build the reconciler by calling the `Build` method on the builder:
   ```go
   reconciler := builder.Build()
   ```

6. Use the built reconciler in your [controller](https://kubernetes.io/docs/concepts/architecture/controller/)
   or [conductor](https://github.com/ethan-gallant/maestro/tree/master/pkg/conductor) to reconcile the child object for
   the parent object.

## Integration with Conductor Package

The Simple Reconciler package seamlessly integrates with
the [Conductor package](https://github.com/ethan-gallant/maestro/tree/master/pkg/conductor) to provide advanced status
condition handling. When used within a conductor, the reconciler automatically updates the status conditions of the
parent object based on the reconciliation result.

To leverage the status condition handling capabilities, follow these steps:

1. Ensure that your parent object has a `Status` field that includes a `Conditions` field of type `[]metav1.Condition`.

2. Register the simple reconciler with the conductor using the `Register` method:
   ```go
   conductor := conductor.ForParent(parent).
       WithClient(client).
       WithContext(ctx).
       WithLogger(logger).
       Build()

   conductor.Register(reconciler)
   ```

3. The simple reconciler will automatically update the status conditions of the parent object based on the
   reconciliation result. It adds the following conditions:
    - `<ReconcilerName>Reconciled`: Indicates whether the reconciliation was successful or not.
    - `<ReconcilerName>Error`: Indicates if an error occurred during reconciliation, along with the error message.

4. You can access the updated status conditions of the parent object in your controller or other reconcilers to make
   decisions based on the reconciliation status.

## Example

Here's a minimal example of how to use the Simple Reconciler package to reconcile a child object for a parent object:

```go
package main

import (
	"context"
	"github.com/ethan-gallant/maestro/pkg/reconciler/simple"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podReconcileFunc(ctx context.Context, deployment *appsv1.Deployment) (*corev1.Pod, error) {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name + "-pod",
			Namespace: deployment.Namespace,
			Labels:    deployment.Spec.Template.Labels,
		},
		Spec: deployment.Spec.Template.Spec,
	}, nil
}

func newPodReconciler() *simple.Reconciler[*appsv1.Deployment, *corev1.Pod] {
	return simple.FromReconcileFunc(podReconcileFunc).
		WithDetails(api.Descriptor{
			Name:        "PodReconciler",
			Description: "Reconciles Pod for Deployment",
		}).
		Build()
}
```

Here's a more detailed example, including custom predicate, deletion, and child key functions:

```go
package main

import (
	"context"
	"github.com/ethan-gallant/maestro/pkg/reconciler/simple"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podReconcileFunc(ctx context.Context, deployment *appsv1.Deployment) (*corev1.Pod, error) {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name + "-pod",
			Namespace: deployment.Namespace,
			Labels:    deployment.Spec.Template.Labels,
		},
		Spec: deployment.Spec.Template.Spec,
	}, nil
}

func podPredicateFunc(deployment *appsv1.Deployment) bool {
	// Only reconcile if the deployment has a specific label
	return deployment.Labels["app"] == "my-app"
}

func podShouldDeleteFunc(deployment *appsv1.Deployment) bool {
	// Delete the pod if the deployment is being deleted
	return !deployment.DeletionTimestamp.IsZero()
}

func podChildKeyFunc(deployment *appsv1.Deployment) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name + "-pod",
			Namespace: deployment.Namespace,
		},
	}
}

func newPodReconciler() *simple.Reconciler[*appsv1.Deployment, *corev1.Pod] {
	return simple.FromReconcileFunc(podReconcileFunc).
		WithDetails(api.Descriptor{
			Name:        "PodReconciler",
			Description: "Reconciles Pod for Deployment",
		}).
		WithPredicateFn(podPredicateFunc).
		WithShouldDeleteFn(podShouldDeleteFunc).
		WithChildKeyFn(podChildKeyFunc).
		Build()
}

func main() {
	reconciler := newPodReconciler()

	// Use the reconciler in your controller or conductor
	// ...
}
```

> 💡 Tip: When your predicate and deletion functions are opposites, you can use the `simple.InversePredicate` function to
> create
> the deletion function from the predicate function. This is useful when you have an `Enabled` field in the parent
> object and want to
> reconcile only when it's `true` and delete the child object when it's `false`.

In this example, the reconciler is created using the `FromReconcileFunc` function, specifying the desired state of the
child object based on the parent object. The reconciler is customized with additional options, such as a predicate
function to control reconciliation, a deletion function to handle child object deletion, and a child key function to
specify the key of the child object. The `newMyReconciler` function encapsulates the builder configuration and returns
the constructed reconciler.

## Best Practices

When using the Simple Reconciler package, consider the following best practices:

1. Keep your reconcile functions focused and single-purpose. Each reconciler should be responsible for reconciling a
   specific child object.

2. Use the predicate function to optimize reconciliation performance by avoiding unnecessary reconciliation cycles.

3. Leverage the dry-run mode for avoiding unnecessary requeues and optimizing performance. Use the `WithDryRunType`
   method with the `DryRunWarn` option to log the differences and set the necessary defaults.

4. Ensure that you set the desired defaults on the objects returned by the reconcile function. If you encounter noisy
   objects that trigger unnecessary updates, consider using the `WithDryRunType` method with the `DryRunWarn` option to
   log the differences and set the necessary defaults.

5. Customize the object comparison options using the `AddCompareOpt` method to avoid unnecessary updates when the child
   object is already in the desired state.

6. Provide meaningful and descriptive names and descriptions for your reconcilers using the `WithDetails` method to
   enhance documentation and debugging.

7. Handle errors gracefully and use the status condition handling capabilities of
   the [Conductor package](https://github.com/ethan-gallant/maestro/tree/master/pkg/conductor) to propagate reconciliation
   status to the parent object.

8. Consider using the `WithShouldDeleteFn` and `WithChildKeyFn` methods to handle the deletion of child objects based on
   custom conditions and to specify the key of the child object.

9. Create separate functions for each builder configuration option and reference them when constructing the reconciler.
   This promotes code reusability and improves readability.

10. Define a function that encapsulates the builder configuration and returns the constructed reconciler. This function
    should be used to create new instances of the reconciler, making it easier to manage and maintain the reconciler
    configuration.

## Additional Resources

For more information on the Kubernetes concepts and resources mentioned in this README, refer to the official Kubernetes
documentation:

- [Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Kubernetes Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
- [Kubernetes Reconciliation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#reconcile)
