package unifiedpushserver

import (
	"context"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestReconcileUnifiedPushServer_Reconcile(t *testing.T) {
	// objects to track in the fake client
	objs := []runtime.Object{
		&pushServerInstance,
	}

	r := buildReconcileWithFakeClientWithMocks(objs, t)

	// mock request to simulate Reconcile() being called on an event for a watched resource
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      pushServerInstance.Name,
			Namespace: pushServerInstance.Namespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check if persistentVolumeClaim has been created
	persistentVolumeClaim := &corev1.PersistentVolumeClaim{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pushServerInstance.Name + "-postgresql", Namespace: pushServerInstance.Namespace}, persistentVolumeClaim)
	if err != nil {
		t.Fatalf("get persistentVolumeClaim: (%v)", err)
	}

	// Check if deployment has been created
	dep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), req.NamespacedName, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	// Check if service has been created
	service := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pushServerInstance.Name + "-postgresql", Namespace: pushServerInstance.Namespace}, service)
	if err != nil {
		t.Fatalf("get service: (%v)", err)
	}

	// Check if serviceAccount has been created
	serviceAccount := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pushServerInstance.Name, Namespace: pushServerInstance.Namespace}, serviceAccount)
	if err != nil {
		t.Fatalf("get serviceAccount: (%v)", err)
	}

	// Check if service has been created
	secret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pushServerInstance.Name + "-postgresql", Namespace: pushServerInstance.Namespace}, secret)
	if err != nil {
		t.Fatalf("get secret: (%v)", err)
	}

	// Check if service has been created
	oauthProxyService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pushServerInstance.Name + "-unifiedpush-proxy", Namespace: pushServerInstance.Namespace}, oauthProxyService)
	if err != nil {
		t.Fatalf("get oauthProxyService: (%v)", err)
	}

	// Check if service has been created
	serviceOauth := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pushServerInstance.Name + "-unifiedpush-proxy", Namespace: pushServerInstance.Namespace}, serviceOauth)
	if err != nil {
		t.Fatalf("get serviceOauth: (%v)", err)
	}

	// Check if route has been created
	serviceOauthRoute := &routev1.Route{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pushServerInstance.Name + "-unifiedpush-proxy", Namespace: pushServerInstance.Namespace}, serviceOauthRoute)
	if err != nil {
		t.Fatalf("get service Oauth route: (%v)", err)
	}

	// Check if servicePush has been created
	servicePush := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pushServerInstance.Name + "-unifiedpush", Namespace: pushServerInstance.Namespace}, servicePush)
	if err != nil {
		t.Fatalf("get servicePush: (%v)", err)
	}

	//TODO:Finish this test

	// Check the result of reconciliation to make sure it has the desired state
	if res.Requeue {
		t.Error("reconcile did requeue which is not expected")
	}
}
