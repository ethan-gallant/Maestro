package reconciler

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var IgnoreManagedFields = cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ManagedFields")
var IgnoreTypeMeta = cmpopts.IgnoreFields(metav1.TypeMeta{}, "APIVersion", "Kind")
