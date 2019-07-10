package androidvariant

import (
	"context"
	"time"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	"github.com/aerogear/unifiedpush-operator/pkg/controller/util"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_androidvariant")

// Add creates a new AndroidVariant Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAndroidVariant{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("androidvariant-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AndroidVariant
	err = c.Watch(&source.Kind{Type: &pushv1alpha1.AndroidVariant{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAndroidVariant{}

// ReconcileAndroidVariant reconciles a AndroidVariant object
type ReconcileAndroidVariant struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AndroidVariant
// object and makes changes based on the state read and what is in the
// AndroidVariant.Spec
// Note:
// The Controller will requeue the Request to be processed again if
// the returned error is non-nil or Result.Requeue is true, otherwise
// upon completion it will remove the work from the queue.
func (r *ReconcileAndroidVariant) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AndroidVariant")

	// Fetch the AndroidVariant instance
	instance := &pushv1alpha1.AndroidVariant{}
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

	// Check if the namespace of the cr is in the APP_NAMESPACES environment variable provided to the operator or if it's in the operator namespace
	operatorNamespace, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return reconcile.Result{}, err
	}
	err = util.IsValidAppNamespace(instance.Namespace, operatorNamespace, instance.Name)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Get a UPS Client for interactions with the UPS service
	unifiedpushClient, err := util.UnifiedpushClient(r.client, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Error getting a UPS Client.", "AndroidVariant.Name", instance.Name)
		return reconcile.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Check if the CR was marked to be deleted
	if instance.GetDeletionTimestamp() != nil {
		// First delete from UPS
		err := unifiedpushClient.DeleteAndroidVariant(instance)
		if err != nil {
			reqLogger.Error(err, "Failed to delete AndroidVariant from UPS", "AndroidVaraint.Name", instance.Name)
			return reconcile.Result{}, err
		}

		// Then unset finalizers
		instance.SetFinalizers(nil)
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	foundVariant, err := unifiedpushClient.GetAndroidVariant(instance)
	if err != nil {
		// this doesn't denote a 404. it is a 500
		reqLogger.Error(err, "Error getting the existing Android variant.", "AndroidVariant.Name", instance.Name)
		return reconcile.Result{}, err
	}

	if foundVariant != "" {
		// we don't do a full reconciliation (update push app on UPS server based on CR content) but we
		// only do initial creation of push apps.
		reqLogger.Info("Skip reconcile: Android Variant already exists in UPS", "AndroidVaraint.Name", instance.Name)
		return reconcile.Result{}, nil
	}

	variantId, err := unifiedpushClient.CreateAndroidVariant(instance)
	if err != nil {
		reqLogger.Error(err, "Error creating Android variant in UPS.", "AndroidVaraint.Name", instance.Name)
		return reconcile.Result{RequeueAfter: time.Second * 5}, nil
	}

	instance.Status.VariantId = variantId
	instance.Status.Ready = true
	err = r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		reqLogger.Error(err, "Error updating AndroidVariant status", "AndroidVariant.Name", instance.Name)
		return reconcile.Result{}, err
	}

	if err := util.AddFinalizer(r.client, reqLogger, instance); err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Android Variant created", "AndroidVariant.Name", instance.Name)
	return reconcile.Result{}, nil
}
