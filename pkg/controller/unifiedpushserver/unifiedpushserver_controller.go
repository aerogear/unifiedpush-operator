package unifiedpushserver

import (
	"context"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	err = c.Watch(&source.Kind{Type: &pushv1alpha1.UnifiedPushServer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Deployment and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Secret and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource PersistentVolumeClaim and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Service and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ServiceAccount and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Route and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
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
	instance := &pushv1alpha1.UnifiedPushServer{}
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

	if instance.Status.Phase == pushv1alpha1.PhaseEmpty {
		instance.Status.Phase = pushv1alpha1.PhaseProvision
		r.client.Status().Update(context.TODO(), instance)
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
	} else {
		desiredImage := postgresql.image()

		containerSpec := findContainerSpec(foundPostgresqlDeployment, POSTGRES_CONTAINER_NAME)
		if containerSpec == nil {
			reqLogger.Info("Skipping image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundPostgresqlDeployment.Namespace, "Deployment.Name", foundPostgresqlDeployment.Name, "ContainerSpec", POSTGRES_CONTAINER_NAME)
		} else if containerSpec.Image != desiredImage {
			reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundPostgresqlDeployment.Namespace, "Deployment.Name", foundPostgresqlDeployment.Name, "ContainerSpec", POSTGRES_CONTAINER_NAME, "ExistingImage", containerSpec.Image, "DesiredImage", desiredImage)

			// update
			updateContainerSpecImage(foundPostgresqlDeployment, POSTGRES_CONTAINER_NAME, desiredImage)

			// enqueue
			err = r.client.Update(context.TODO(), foundPostgresqlDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundPostgresqlDeployment.Namespace, "Deployment.Name", foundPostgresqlDeployment.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
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

	serviceAccount, err := newUnifiedPushServiceAccount(instance)

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

	unifiedpushService, err := newUnifiedPushServerService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, unifiedpushService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundUnifiedpushService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: unifiedpushService.Name, Namespace: unifiedpushService.Namespace}, foundUnifiedpushService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", unifiedpushService.Namespace, "Service.Name", unifiedpushService.Name)
		err = r.client.Create(context.TODO(), unifiedpushService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	oauthProxyRoute, err := newOauthProxyRoute(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, oauthProxyRoute, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Route already exists
	foundOauthProxyRoute := &routev1.Route{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: oauthProxyRoute.Name, Namespace: oauthProxyRoute.Namespace}, foundOauthProxyRoute)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Route", "Route.Namespace", oauthProxyRoute.Namespace, "Route.Name", oauthProxyRoute.Name)
		err = r.client.Create(context.TODO(), oauthProxyRoute)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new Deployment object
	unifiedpushDeployment, err := newUnifiedPushServerDeployment(instance)

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
	} else {
		desiredUnifiedPushImage := unifiedpush.image()
		desiredProxyImage := proxy.image()

		unifiedPushContainerSpec := findContainerSpec(foundUnifiedpushDeployment, UPS_CONTAINER_NAME)
		if unifiedPushContainerSpec == nil {
			reqLogger.Info("Skipping image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "ContainerSpec", UPS_CONTAINER_NAME)
		} else if unifiedPushContainerSpec.Image != desiredUnifiedPushImage {
			reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "ContainerSpec", UPS_CONTAINER_NAME, "ExistingImage", unifiedPushContainerSpec.Image, "DesiredImage", desiredUnifiedPushImage)

			// update
			updateContainerSpecImage(foundUnifiedpushDeployment, UPS_CONTAINER_NAME, desiredUnifiedPushImage)

			// enqueue
			err = r.client.Update(context.TODO(), foundUnifiedpushDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}

		proxyContainerSpec := findContainerSpec(foundUnifiedpushDeployment, OAUTH_PROXY_CONTAINER_NAME)
		if proxyContainerSpec == nil {
			reqLogger.Info("Skipping image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "ContainerSpec", OAUTH_PROXY_CONTAINER_NAME)
		} else if proxyContainerSpec.Image != desiredProxyImage {
			reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "ContainerSpec", OAUTH_PROXY_CONTAINER_NAME, "ExistingImage", proxyContainerSpec.Image, "DesiredImage", desiredProxyImage)

			// update
			updateContainerSpecImage(foundUnifiedpushDeployment, OAUTH_PROXY_CONTAINER_NAME, desiredProxyImage)

			// enqueue
			err = r.client.Update(context.TODO(), foundUnifiedpushDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	}

	if foundUnifiedpushDeployment.Status.ReadyReplicas > 0 && instance.Status.Phase != pushv1alpha1.PhaseComplete {
		instance.Status.Phase = pushv1alpha1.PhaseComplete
		r.client.Status().Update(context.TODO(), instance)
	}

	// Resources already exist - don't requeue
	reqLogger.Info("Skip reconcile: Resources already exist")
	return reconcile.Result{}, nil
}
