package reconciler

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var ErrChildKeyMismatch = errors.New("child key mismatch")

func InvertFunc[T client.Object](f func(parent T) bool) func(parent T) bool {
	return func(parent T) bool {
		return !f(parent)
	}
}

func IgnoreAnnotations() cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		return p.String() == "ObjectMeta.Annotations"
	}, cmp.Ignore())
}

func IgnoreFinalizers() cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		return p.String() == "ObjectMeta.Finalizers"
	}, cmp.Ignore())
}

func IgnoreStatusFields() cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		return p.String() == "Status" || strings.HasPrefix(p.String(), "Status.")
	}, cmp.Ignore())
}

func IgnoreManagedFields() cmp.Option {
	return cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ManagedFields")
}

func IgnoreTypeMeta() cmp.Option {
	return cmpopts.IgnoreFields(metav1.TypeMeta{}, "APIVersion", "Kind")
}
