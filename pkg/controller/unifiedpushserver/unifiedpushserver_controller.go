package unifiedpushserver

import (
	"context"

	aerogearv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/aerogear/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_unifiedpushserver")

// Add creates a new UnifiedPushServer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileUnifiedPushServer{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("unifiedpushserver-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &aerogearv1alpha1.UnifiedPushServer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployment and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &aerogearv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Secret and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &aerogearv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource PersistentVolumeClaim and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &aerogearv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Service and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &aerogearv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ServiceAccount and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &aerogearv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RoleBinding and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &aerogearv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileUnifiedPushServer{}

// ReconcileUnifiedPushServer reconciles a UnifiedPushServer object
type ReconcileUnifiedPushServer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a UnifiedPushServer object and makes changes based on the state read
// and what is in the UnifiedPushServer.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileUnifiedPushServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling UnifiedPushServer")

	// Fetch the UnifiedPushServer instance
	instance := &aerogearv1alpha1.UnifiedPushServer{}
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

	persistentVolumeClaim, err := newPostgresqlPersistentVolumeClaim(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, persistentVolumeClaim, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this PersistentVolumeClaim already exists
	foundPersistentVolumeClaim := &corev1.PersistentVolumeClaim{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: persistentVolumeClaim.Name, Namespace: persistentVolumeClaim.Namespace}, foundPersistentVolumeClaim)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", persistentVolumeClaim.Namespace, "PersistentVolumeClaim.Name", persistentVolumeClaim.Name)
		err = r.client.Create(context.TODO(), persistentVolumeClaim)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	postgresqlDeployment, err := newPostgresqlDeployment(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, postgresqlDeployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists
	foundPostgresqlDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: postgresqlDeployment.Name, Namespace: postgresqlDeployment.Namespace}, foundPostgresqlDeployment)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", postgresqlDeployment.Namespace, "Deployment.Name", postgresqlDeployment.Name)
		err = r.client.Create(context.TODO(), postgresqlDeployment)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	postgresqlService, err := newPostgresqlService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, postgresqlService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundPostgresqlService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: postgresqlService.Name, Namespace: postgresqlService.Namespace}, foundPostgresqlService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", postgresqlService.Namespace, "Service.Name", postgresqlService.Name)
		err = r.client.Create(context.TODO(), postgresqlService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	serviceAccount := newOauthProxyServiceAccount(instance)

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, serviceAccount, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ServiceAccount already exists
	foundServiceAccount := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
		err = r.client.Create(context.TODO(), serviceAccount)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	roleBinding := newOauthProxySaRoleBinding(instance)

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, roleBinding, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ServiceAccount already exists
	foundRoleBinding := &rbacv1.RoleBinding{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, foundRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new RoleBinding", "RoleBinding.Namespace", roleBinding.Namespace, "RoleBinding.Name", roleBinding.Name)
		err = r.client.Create(context.TODO(), roleBinding)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	postgresqlSecret, err := newPostgresqlSecret(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, postgresqlSecret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Secret already exists
	foundPostgresqlSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: postgresqlSecret.Name, Namespace: postgresqlSecret.Namespace}, foundPostgresqlSecret)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", postgresqlSecret.Namespace, "Secret.Name", postgresqlSecret.Name)
		err = r.client.Create(context.TODO(), postgresqlSecret)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	oauthProxyService, err := newOauthProxyService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, oauthProxyService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundOauthProxyService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: oauthProxyService.Name, Namespace: oauthProxyService.Namespace}, foundOauthProxyService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", oauthProxyService.Namespace, "Service.Name", oauthProxyService.Name)
		err = r.client.Create(context.TODO(), oauthProxyService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new Deployment object
	unifiedpushDeployment := newUnifiedPushServerDeployment(instance)

	if err := controllerutil.SetControllerReference(instance, unifiedpushDeployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists
	foundUnifiedpushDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: unifiedpushDeployment.Name, Namespace: unifiedpushDeployment.Namespace}, foundUnifiedpushDeployment)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", unifiedpushDeployment.Namespace, "Deployment.Name", unifiedpushDeployment.Name)
		err = r.client.Create(context.TODO(), unifiedpushDeployment)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Resources already exist - don't requeue
	reqLogger.Info("Skip reconcile: Resources already exist")
	return reconcile.Result{}, nil
}
