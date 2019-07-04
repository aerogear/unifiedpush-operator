package androidvariant

import (
	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

//buildReconcileWithFakeClientWithMocks return reconcile with fake client, schemes and mock objects
func buildReconcileWithFakeClientWithMocks(objs []runtime.Object, t *testing.T) *ReconcileAndroidVariant {
	s := scheme.Scheme

	// Add route Openshift scheme
	if err := routev1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add route scheme: (%v)", err)
	}

	s.AddKnownTypes(pushv1alpha1.SchemeGroupVersion, &pushv1alpha1.AndroidVariant{})

	// create a fake client to mock API calls with the mock objects
	cl := fake.NewFakeClient(objs...)

	// create a ReconcileMobileSecurityService object with the scheme and fake client
	return &ReconcileAndroidVariant{client: cl, scheme: s}
}
