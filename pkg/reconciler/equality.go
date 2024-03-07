package reconciler

import (
	"errors"

	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IgnoreManagedFields = cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ManagedFields")
var IgnoreTypeMeta = cmpopts.IgnoreFields(metav1.TypeMeta{}, "APIVersion", "Kind")

var ErrChildKeyMismatch = errors.New("child key mismatch")

func InvertFunc[T client.Object](f func(parent T) bool) func(parent T) bool {
	return func(parent T) bool {
		return !f(parent)
	}
}
