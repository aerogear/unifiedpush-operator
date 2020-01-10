package unifiedpushserver

import (
	"context"
	"fmt"
	"testing"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"

	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
		{
			name:  "should create expected resources on reconcile of cr with external DB details",
			given: &crWithExternalDatabase,
			expect: map[string]runtime.Object{
				crWithExternalDatabase.Name:                                      &appsv1.Deployment{},
				crWithExternalDatabase.Name:                                      &corev1.ServiceAccount{},
				fmt.Sprintf("%s-postgresql", crWithExternalDatabase.Name):        &corev1.Secret{},
				fmt.Sprintf("%s-unifiedpush", crWithExternalDatabase.Name):       &corev1.Service{},
				fmt.Sprintf("%s-unifiedpush-proxy", crWithExternalDatabase.Name): &corev1.Service{},
				fmt.Sprintf("%s-unifiedpush-proxy", crWithExternalDatabase.Name): &routev1.Route{},
			},
		},
		{
			name:  "should create expected resources on reconcile of cr with one backup specified",
			given: &crWithBackup,
			expect: map[string]runtime.Object{
				crWithBackup.Name:  &appsv1.Deployment{},
				crWithBackup.Name:  &corev1.ServiceAccount{},
				"example-backup-1": &batchv1beta1.CronJob{},
				"example-backup-2": &batchv1beta1.CronJob{},
				fmt.Sprintf("%s-postgresql", crWithBackup.Name):        &corev1.PersistentVolumeClaim{},
				fmt.Sprintf("%s-postgresql", crWithBackup.Name):        &corev1.Service{},
				fmt.Sprintf("%s-postgresql", crWithBackup.Name):        &corev1.Secret{},
				fmt.Sprintf("%s-unifiedpush", crWithBackup.Name):       &corev1.Service{},
				fmt.Sprintf("%s-unifiedpush-proxy", crWithBackup.Name): &corev1.Service{},
				fmt.Sprintf("%s-unifiedpush-proxy", crWithBackup.Name): &routev1.Route{},
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
	crWithExternalDatabase = pushv1alpha1.UnifiedPushServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-with-external-db",
			Namespace: "unifiedpush",
		},
		Spec: pushv1alpha1.UnifiedPushServerSpec{
			ExternalDB: true,
			Database: pushv1alpha1.UnifiedPushServerDatabase{
				Name:     "test",
				User:     "me",
				Password: "password",
				Host:     "127.0.0.1",
				Port:     intstr.FromInt(5432),
			},
		},
	}
	crWithBackup = pushv1alpha1.UnifiedPushServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-with-backups",
			Namespace: "unifiedpush",
		},
		Spec: pushv1alpha1.UnifiedPushServerSpec{
			Backups: []pushv1alpha1.UnifiedPushServerBackup{
				pushv1alpha1.UnifiedPushServerBackup{
					Name:                   "example-backup-1",
					Schedule:               "0 0 0 0 0",
					BackendSecretName:      "example-with-backup-postgresql",
					BackendSecretNamespace: "unifiedpush",
				},
				pushv1alpha1.UnifiedPushServerBackup{
					Name:                   "example-backup-2",
					Schedule:               "0 0 0 0 0",
					BackendSecretName:      "example-with-backup-postgresql",
					BackendSecretNamespace: "unifiedpush",
				},
			},
		},
	}
)
