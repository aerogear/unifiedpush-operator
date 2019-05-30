package pushapplication

import (
	"context"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"

	"github.com/aerogear/unifiedpush-operator/pkg/unifiedpush"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_pushapplication")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PushApplication Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePushApplication{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pushapplication-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource PushApplication
	err = c.Watch(&source.Kind{Type: &pushv1alpha1.PushApplication{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePushApplication{}

// ReconcilePushApplication reconciles a PushApplication object
type ReconcilePushApplication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PushApplication object and makes changes based on the state read
// and what is in the PushApplication.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePushApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling PushApplication")

	// Fetch the PushApplication instance
	instance := &pushv1alpha1.PushApplication{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO: aliok + grdryn
	upsClient := unifiedpush.UnifiedpushClient{"http://example-unifiedpushserver-unifiedpush-unifiedpush.192.168.42.227.nip.io"}

	// TODO: aliok + grdryn
	// Check if this push app already exists
	pushAppName := instance.Name
	foundApp, err := upsClient.GetApplication(pushAppName)

	if err != nil {
		// this doesn't denote a 404. it is a 500
		reqLogger.Error(err, "Error getting the existing push application.", "PushApp.Name", pushAppName)
		return reconcile.Result{}, err
	}

	if foundApp != nil {
		// we don't do a full reconciliation (update push app on UPS server based on CR content) but we
		// only do initial creation of push apps.
		reqLogger.Info("Skip reconcile: Push app on UPS already exists", "PushApp.Name", pushAppName)
		return reconcile.Result{}, nil
	}

	pushApp := pushv1alpha1.PushApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name: pushAppName,
		},
		Spec: pushv1alpha1.PushApplicationSpec{
			Description: instance.Spec.Description,
		},
	}

	appId, secret, err := upsClient.CreateApplication(&pushApp)
	if err != nil {
		reqLogger.Error(err, "Error creating push application.", "PushApp.Name", pushAppName)
		return reconcile.Result{}, err
	}
	reqLogger.Info("Push app created", "PushApp.Name", pushAppName, "PushApp.Id", appId, "PushApp.MasterSecret", secret)
	return reconcile.Result{}, nil
}
