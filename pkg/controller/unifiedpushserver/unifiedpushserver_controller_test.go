package unifiedpushserver

import (
	"context"
	"fmt"
	"testing"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"

	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileUnifiedPushServer_Reconcile(t *testing.T) {
	scenarios := []struct {
		name   string
		given  *pushv1alpha1.UnifiedPushServer
		expect map[string]runtime.Object
	}{
		{
			name:  "should create expected resources on reconcile of simple empty cr",
			given: &crWithDefaults,
			expect: map[string]runtime.Object{
				crWithDefaults.Name: &appsv1.Deployment{},
				crWithDefaults.Name: &corev1.ServiceAccount{},
				fmt.Sprintf("%s-postgresql", crWithDefaults.Name):        &corev1.PersistentVolumeClaim{},
				fmt.Sprintf("%s-postgresql", crWithDefaults.Name):        &corev1.Service{},
				fmt.Sprintf("%s-postgresql", crWithDefaults.Name):        &corev1.Secret{},
				fmt.Sprintf("%s-unifiedpush", crWithDefaults.Name):       &corev1.Service{},
				fmt.Sprintf("%s-unifiedpush-proxy", crWithDefaults.Name): &corev1.Service{},
				fmt.Sprintf("%s-unifiedpush-proxy", crWithDefaults.Name): &routev1.Route{},
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {

			// given
			r := buildReconcileWithFakeClientWithMocks([]runtime.Object{scenario.given}, t)
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      scenario.given.Name,
					Namespace: scenario.given.Namespace,
				},
			}

			// Required "backupjob" SA for backup CronJobs
			err := r.client.Create(context.TODO(), &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "backupjob", Namespace: "unifiedpush"}})

			// when
			res, err := r.Reconcile(req)
			if err != nil {
				t.Fatalf("reconcile: (%v)", err)
			}
			if res.Requeue {
				t.Error("Reconcile requeued unexpectedly")
			}

			// then
			for n, o := range scenario.expect {
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: n, Namespace: scenario.given.Namespace}, o)
				if err != nil {
					t.Fatalf("get %s %s: (%v)", o.GetObjectKind(), n, err)
				}
			}
		})
	}
}

var (
	crWithDefaults = pushv1alpha1.UnifiedPushServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-unifiedpushserver",
			Namespace: "unifiedpush",
		},
	}
)
