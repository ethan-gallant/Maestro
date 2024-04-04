# Maestro 🎼

[![Go Report Card](https://goreportcard.com/badge/github.com/ethan-gallant/maestro)](https://goreportcard.com/report/github.com/ethan-gallant/maestro)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Maestro is an opinionated framework designed to streamline the orchestration of reconcilers within the Kubernetes
ecosystem. By adopting a clear and structured approach, Maestro simplifies the registration, execution, and management
of reconcilers, enabling developers to focus on the core logic of their applications without getting bogged down by the
intricacies of Kubernetes resource management. 🚀

## Features ✨

- 🎉 Simple and intuitive API for building reconcilers
- 🔍 Supports generic parent and child objects that implement
  the [`client.Object`](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#Object) interface
- 🏗️ Provides a fluent builder pattern for constructing reconcilers with ease
- 🎯 Offers customizable predicate functions to control reconciliation behavior
- 🔗 Supports optional owner references for child objects
- 🏃‍♂️ Enables dry-run mode for avoiding unnecessary requeues and optimizing performance
- 🔄 Allows customization of object comparison options to avoid unnecessary updates
- 🗑️ Supports deletion of child objects based on custom conditions
- 🔧 Integrates with the [Conductor package](https://github.com/ethan-gallant/maestro/tree/master/pkg/conductor) for
  advanced status condition handling

## Getting Started 🌟

### Prerequisites

- Go 1.18 or higher
- Kubernetes cluster
- [`controller-runtime`](https://github.com/kubernetes-sigs/controller-runtime) library

### Installation

To install Maestro, use the following command:

```shell
go get github.com/ethan-gallant/maestro
```

### Usage

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

For more detailed examples and usage instructions, please refer to
the [Simple Reconciler Package README](https://github.com/ethan-gallant/maestro/tree/master/pkg/reconciler/simple).

## Key Components 🔑

### Conductor 🎛

The Conductor is a generic interface that facilitates the registration and coordination of reconcilers. It acts as the
central hub through which all reconcilers are managed. For more information, see
the [Conductor Package README](https://github.com/ethan-gallant/maestro/tree/master/pkg/conductor).

### Reconciler 🔄

The Reconciler is an interface defining the essential method `Reconcile`, which encapsulates the logic for reconciling
the state of a Kubernetes resource. Maestro provides a Simple Reconciler implementation in
the [Simple Reconciler Package](https://github.com/ethan-gallant/maestro/tree/master/pkg/reconciler/simple).

### Binder 🔗

The Binder package provides a convenient way to bind and retrieve values from a `context.Context` using static or
dynamic keys. It offers a type-safe and flexible approach to store and access custom data within a context. For more
information, see the [Binder Package README](https://github.com/ethan-gallant/maestro/tree/master/pkg/binder).

## Documentation 📚

For detailed documentation and examples, please refer to the individual package READMEs:

- [Conductor Package](https://github.com/ethan-gallant/maestro/tree/master/pkg/conductor)
- [Simple Reconciler Package](https://github.com/ethan-gallant/maestro/tree/master/pkg/reconciler/simple)
- [Binder Package](https://github.com/ethan-gallant/maestro/tree/master/pkg/binder)

## Contributing 🤝

We welcome contributions to Maestro! If you have any ideas, suggestions, or bug reports, please open an issue or submit
a pull request. For more information, see the [Contributing Guide](docs/CONTRIBUTING.md).

## License 📜

Maestro is released under the [MIT License](LICENSE.md).

Happy orchestrating with Maestro! 🎉