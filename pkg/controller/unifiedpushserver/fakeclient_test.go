package unifiedpushserver

import (
	"testing"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	enmassev1beta1 "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1beta1"
	messaginguserv1beta1 "github.com/enmasseproject/enmasse/pkg/apis/user/v1beta1"
	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

//buildReconcileWithFakeClientWithMocks return reconcile with fake client, schemes and mock objects
func buildReconcileWithFakeClientWithMocks(objs []runtime.Object, t *testing.T) *ReconcileUnifiedPushServer {
	s := scheme.Scheme

	// Add Openshift route scheme
	if err := routev1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add route scheme: (%v)", err)
	}

	// Add Prometheus monitoring scheme
	if err := monitoringv1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add monitoringv1 scheme: (%v)", err)
	}

	// Add integreatly scheme
	if err := integreatlyv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add monitoringv1 scheme: (%v)", err)
	}

	// Add enmasse scheme
	if err := enmassev1beta1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add monitoringv1 scheme: (%v)", err)
	}

	// Add enmasse user scheme
	if err := messaginguserv1beta1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add monitoringv1 scheme: (%v)", err)
	}

	s.AddKnownTypes(pushv1alpha1.SchemeGroupVersion, &pushv1alpha1.UnifiedPushServer{})
	s.AddKnownTypes(pushv1alpha1.SchemeGroupVersion, &pushv1alpha1.UnifiedPushServerList{})

	// create a fake client to mock API calls with the mock objects
	cl := fake.NewFakeClient(objs...)

	fakeApiVersionChecker := &apiVersionChecker{
		check: func(apiGroupVersion string) (bool, error) { return true, nil },
	}

	return &ReconcileUnifiedPushServer{client: cl, scheme: s, apiVersionChecker: fakeApiVersionChecker}
}
