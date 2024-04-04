package reconciler

import "sigs.k8s.io/controller-runtime/pkg/client"

func IsNotMarkedForDeletion[T client.Object](obj T) bool {
	return obj.GetDeletionTimestamp() == nil
}
