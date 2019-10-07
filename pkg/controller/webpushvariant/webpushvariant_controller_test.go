package webpushvariant

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileWebPushVariant_Reconcile(t *testing.T) {
	//TODO: Add just as example for we start to cover it with unit tests.
	// objects to track in the fake client
	objs := []runtime.Object{
		&webpushVariantInstance,
	}

	r := buildReconcileWithFakeClientWithMocks(objs, t)

	// mock request to simulate Reconcile() being called on an event for a watched resource
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      webpushVariantInstance.Name,
			Namespace: webpushVariantInstance.Namespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state
	if res.Requeue {
		t.Error("reconcile did requeue which is not expected")
	}
}
