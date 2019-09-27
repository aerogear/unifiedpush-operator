package unifiedpushserver

import (
	"context"
	"github.com/aerogear/unifiedpush-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"time"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	"github.com/aerogear/unifiedpush-operator/pkg/config"
	"github.com/aerogear/unifiedpush-operator/pkg/nspredicate"

	enmassev1beta "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1beta1"
	messaginguserv1beta "github.com/enmasseproject/enmasse/pkg/apis/user/v1beta1"

	openshiftappsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/client-go/rest"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

var (
	cfg = config.New()
	log = logf.Log.WithName("controller_unifiedpushserver")
)

// Add creates a new UnifiedPushServer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileUnifiedPushServer{config: mgr.GetConfig(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("unifiedpushserver-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource UnifiedPushServer
	onlyEnqueueForServiceNamespace, err := nspredicate.NewFromEnvVar("SERVICE_NAMESPACE")
	if err != nil {
		return err
	}
	err = c.Watch(
		&source.Kind{Type: &pushv1alpha1.UnifiedPushServer{}},
		&handler.EnqueueRequestForObject{},
		onlyEnqueueForServiceNamespace,
	)
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource DeploymentConfig and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &openshiftappsv1.DeploymentConfig{}}, &handler.EnqueueRequestForOwner{
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

	// Watch for changes to secondary resource CronJob and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &batchv1beta1.CronJob{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource MessagingUser and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &messaginguserv1beta.MessagingUser{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	// If the problem is just a missing kind, don't worry about it
	if _, isNoKindMatchError := err.(*meta.NoKindMatchError); err != nil && !isNoKindMatchError {
		return err
	}

	// Watch for changes to secondary resource AddressSpace and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &enmassev1beta.AddressSpace{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	// If the problem is just a missing kind, don't worry about it
	if _, isNoKindMatchError := err.(*meta.NoKindMatchError); err != nil && !isNoKindMatchError {
		return err
	}

	// Watch for changes to secondary resource Address and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &enmassev1beta.Address{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	// If the problem is just a missing kind, don't worry about it
	if _, isNoKindMatchError := err.(*meta.NoKindMatchError); err != nil && !isNoKindMatchError {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileUnifiedPushServer{}

// ReconcileUnifiedPushServer reconciles a UnifiedPushServer object
type ReconcileUnifiedPushServer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	config *rest.Config
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

	//Create new client; we want to avoid caching the enmasse resources we watch.
	operatorClient, err := client.New(r.config, client.Options{})
	if err != nil {
		return reconcile.Result{}, err
	}
	// Fetch the UnifiedPushServer instance
	instance := &pushv1alpha1.UnifiedPushServer{}
	err = operatorClient.Get(context.TODO(), request.NamespacedName, instance)
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

	// look for other unifiedPush resources and don't provision a new one if there is another one with Phase=Complete
	existingInstances := &pushv1alpha1.UnifiedPushServerList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UnifiedPushServer",
			APIVersion: "push.aerogear.org/v1alpha1",
		},
	}
	opts := &client.ListOptions{Namespace: instance.Namespace}
	err = operatorClient.List(context.TODO(), opts, existingInstances)
	if err != nil {
		reqLogger.Error(err, "Failed to list UnifiedPush resources", "UnifiedPush.Namespace", instance.Namespace)
		return reconcile.Result{}, err
	} else if len(existingInstances.Items) > 1 { // check if > 1 since there's the current one already in that list.
		for _, existingInstance := range existingInstances.Items {
			if existingInstance.Name == instance.Name {
				continue
			}
			if existingInstance.Status.Phase == pushv1alpha1.PhaseProvision || existingInstance.Status.Phase == pushv1alpha1.PhaseComplete {
				reqLogger.Info("There is already a UnifiedPush resource in Complete phase. Doing nothing for this CR.", "UnifiedPush.Namespace", instance.Namespace, "UnifiedPush.Name", instance.Name)
				return reconcile.Result{}, nil
			}
		}
	}

	if instance.Status.Phase == pushv1alpha1.PhaseEmpty {
		instance.Status.Phase = pushv1alpha1.PhaseProvision
		err = operatorClient.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update UnifiedPush resource status phase", "UnifiedPush.Namespace", instance.Namespace, "UnifiedPush.Name", instance.Name)
			return reconcile.Result{}, err
		}
	}

	//#region AMQ resource reconcile
	if instance.Spec.UseMessageBroker {
		//#region create addressSpace
		addressSpace := newAddressSpace(instance)

		// Set UnifiedPushServer instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, addressSpace, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		foundAddressSpace := &enmassev1beta.AddressSpace{}
		err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: addressSpace.Name, Namespace: addressSpace.Namespace}, foundAddressSpace)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Address Space", "AddressSpace.Namespace", addressSpace.Namespace, "AddressSpace.Name", addressSpace.Name)
			err = operatorClient.Create(context.TODO(), addressSpace)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("Requeuing, AddressSpace not ready.", "AddressSpace.Namespace", addressSpace.Namespace, "AddressSpace.Name", addressSpace.Name)
			return reconcile.Result{RequeueAfter: time.Second * 10}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		} else if !foundAddressSpace.Status.IsReady {
			reqLogger.Info("Requeuing, AddressSpace not ready.", "AddressSpace.Namespace", foundAddressSpace.Namespace, "AddressSpace.Name", foundAddressSpace.Name)
			return reconcile.Result{RequeueAfter: time.Second * 10}, nil
		} else {
			reqLogger.Info("Found AddressSpace for UPS")
		}
		//#endregion

		//#region check that user exists
		user, err := newMessagingUser(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Set UnifiedPushServer instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, user, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		foundUser := &messaginguserv1beta.MessagingUser{}
		err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: user.Name, Namespace: user.Namespace}, foundUser)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new MessagingUser", "MessagingUser.Namespace", user.Namespace, "MessagingUser.Name", user.Name)
			err = operatorClient.Create(context.TODO(), user)
			if err != nil {
				return reconcile.Result{}, err
			}

		} else if err != nil {
			return reconcile.Result{}, err
		}
		//#endregion

		//#region create secret for user password and artemis url
		for _, status := range foundAddressSpace.Status.EndpointStatus {
			if status.Name == "messaging" { //"messaging" is a key from enmasse.
				addressSpaceURL := status.ServiceHost
				password := string(user.Spec.Authentication.Password)
				secret := newAMQSecret(instance, password, addressSpaceURL)
				foundSecret := &corev1.Secret{}
				err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, foundSecret)
				if err != nil && errors.IsNotFound(err) {
					reqLogger.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
					err = operatorClient.Create(context.TODO(), secret)
					if err != nil {
						return reconcile.Result{}, err
					}
				} else if err != nil {
					return reconcile.Result{}, err
				}
				break
			}
		}
		//#endregion

		//#region queues
		queues := []string{"APNsPushMessageQueue", "APNsTokenBatchQueue", "GCMPushMessageQueue", "GCMTokenBatchQueue", "WNSPushMessageQueue", "WNSTokenBatchQueue", "MetricsQueue", "TriggerMetricCollectionQueue", "TriggerVariantMetricCollectionQueue", "BatchLoadedQueue", "AllBatchesLoadedQueue", "FreeServiceSlotQueue"}
		requeueCreate := false
		for _, address := range queues {
			queue := newQueue(instance, address)
			foundQueue := &enmassev1beta.Address{}
			// Set UnifiedPushServer instance as the owner and controller
			if err := controllerutil.SetControllerReference(instance, queue, r.scheme); err != nil {
				return reconcile.Result{}, err
			}

			err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, foundQueue)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Queue", "Queue.Namespace", queue.Namespace, "Queue.Name", queue.Name)
				err = operatorClient.Create(context.TODO(), queue)
				if err != nil {
					return reconcile.Result{}, err
				}
				requeueCreate = true
			} else if err != nil {
				reqLogger.Info("Queue Error")
				return reconcile.Result{}, err
			} else if !foundQueue.Status.IsReady {
				reqLogger.Info("Queue Not ready", "Queue.Name", foundQueue.Name)
				requeueCreate = true
			}
		}

		if requeueCreate {
			reqLogger.Info("Requeueing while queues are created")
			return reconcile.Result{RequeueAfter: time.Second * 5}, nil
		}
		//#endregion

		reqLogger.Info("Found all queues  for UPS")

		//#region topics
		topics := []string{"MetricsProcessingStartedTopic", "topic/APNSClient"}
		for _, address := range topics {
			topic := newTopic(instance, address)
			foundTopic := &enmassev1beta.Address{}
			// Set UnifiedPushServer instance as the owner and controller
			if err := controllerutil.SetControllerReference(instance, topic, r.scheme); err != nil {
				return reconcile.Result{}, err
			}

			err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: topic.Name, Namespace: topic.Namespace}, foundTopic)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Topic", "Topic.Namespace", topic.Namespace, "Topic.Name", topic.Name)
				err = operatorClient.Create(context.TODO(), topic)
				if err != nil {
					return reconcile.Result{}, err
				}
				requeueCreate = true
			} else if err != nil {
				return reconcile.Result{}, err
			}
		}

		if requeueCreate {
			reqLogger.Info("Requeueing while topics are created")
			return reconcile.Result{RequeueAfter: time.Second * 5}, nil
		}
		//#endregion

		reqLogger.Info("Found All queues and topics for UPS")

	}
	//#endregion

	//#region Postgres PVC
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
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: persistentVolumeClaim.Name, Namespace: persistentVolumeClaim.Namespace}, foundPersistentVolumeClaim)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", persistentVolumeClaim.Namespace, "PersistentVolumeClaim.Name", persistentVolumeClaim.Name)
		err = operatorClient.Create(context.TODO(), persistentVolumeClaim)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		requiredPostgresPVCSize := getPostgresPVCSize(instance)

		foundPVCSize := foundPersistentVolumeClaim.Spec.Resources.Requests[corev1.ResourceStorage]
		if foundPVCSize.String() != requiredPostgresPVCSize {
			reqLogger.Info("Request size of PersistentVolumeClaim is different than in the UnifiedPushServer spec or the operator defaults", "PersistentVolumeClaim.Namespace", persistentVolumeClaim.Namespace, "PersistentVolumeClaim.Name", persistentVolumeClaim.Name, "Found size", foundPVCSize.String(), "Spec size", requiredPostgresPVCSize)

			foundPersistentVolumeClaim.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(requiredPostgresPVCSize)

			// enqueue
			err = operatorClient.Update(context.TODO(), foundPersistentVolumeClaim)
			if err != nil {
				reqLogger.Error(err, "Failed to update PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", foundPersistentVolumeClaim.Namespace, "PersistentVolumeClaim.Name", foundPersistentVolumeClaim.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}

	}
	//#endregion

	//#region Postgres DeploymentConfig
	postgresqlDeploymentConfig, err := newPostgresqlDeploymentConfig(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, postgresqlDeploymentConfig, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DeploymentConfig already exists
	foundPostgresqlDeploymentConfig := &openshiftappsv1.DeploymentConfig{}
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: postgresqlDeploymentConfig.Name, Namespace: postgresqlDeploymentConfig.Namespace}, foundPostgresqlDeploymentConfig)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new DeploymentConfig", "DeploymentConfig.Namespace", postgresqlDeploymentConfig.Namespace, "DeploymentConfig.Name", postgresqlDeploymentConfig.Name)
		err = operatorClient.Create(context.TODO(), postgresqlDeploymentConfig)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		postgresResourceRequirements := getPostgresResourceRequirements(instance)

		containers := foundPostgresqlDeploymentConfig.Spec.Template.Spec.Containers
		for i := range containers {
			if containers[i].Name == cfg.PostgresContainerName {
				if reflect.DeepEqual(containers[i].Resources, postgresResourceRequirements) == false {
					reqLogger.Info("Postgres container resource requirements are different than in the UnifiedPushServer spec or the operator defaults", "DeploymentConfig.Namespace", foundPostgresqlDeploymentConfig.Namespace, "DeploymentConfig.Name", foundPostgresqlDeploymentConfig.Name, "Found resource requirements", containers[i].Resources, "Spec resource requirements", postgresResourceRequirements)

					containers[i].Resources = postgresResourceRequirements

					// enqueue
					err = operatorClient.Update(context.TODO(), foundPostgresqlDeploymentConfig)
					if err != nil {
						reqLogger.Error(err, "Failed to update DeploymentConfig", "DeploymentConfig.Namespace", foundPostgresqlDeploymentConfig.Namespace, "DeploymentConfig.Name", foundPostgresqlDeploymentConfig.Name)
						return reconcile.Result{}, err
					}
					return reconcile.Result{Requeue: true}, nil
				}
			}
		}

		desiredImage := constants.PostgresImage

		containerSpec := findContainerSpec(foundPostgresqlDeploymentConfig, cfg.PostgresContainerName)
		if containerSpec == nil {
			reqLogger.Info("Skipping image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundPostgresqlDeploymentConfig.Namespace, "Deployment.Name", foundPostgresqlDeploymentConfig.Name, "ContainerSpec", cfg.PostgresContainerName)
		} else if containerSpec.Image != desiredImage {
			reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundPostgresqlDeploymentConfig.Namespace, "Deployment.Name", foundPostgresqlDeploymentConfig.Name, "ContainerSpec", cfg.PostgresContainerName, "ExistingImage", containerSpec.Image, "DesiredImage", desiredImage)

			// update
			updateContainerSpecImage(foundPostgresqlDeploymentConfig, cfg.PostgresContainerName, desiredImage)

			// enqueue
			err = operatorClient.Update(context.TODO(), foundPostgresqlDeploymentConfig)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundPostgresqlDeploymentConfig.Namespace, "Deployment.Name", foundPostgresqlDeploymentConfig.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	}
	//#endregion

	//#region Postgres Service
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
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: postgresqlService.Name, Namespace: postgresqlService.Namespace}, foundPostgresqlService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", postgresqlService.Namespace, "Service.Name", postgresqlService.Name)
		err = operatorClient.Create(context.TODO(), postgresqlService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region ServiceAccount
	serviceAccount, err := newUnifiedPushServiceAccount(instance)

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, serviceAccount, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ServiceAccount already exists
	foundServiceAccount := &corev1.ServiceAccount{}
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
		err = operatorClient.Create(context.TODO(), serviceAccount)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region Postgres Secret
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
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: postgresqlSecret.Name, Namespace: postgresqlSecret.Namespace}, foundPostgresqlSecret)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", postgresqlSecret.Namespace, "Secret.Name", postgresqlSecret.Name)
		err = operatorClient.Create(context.TODO(), postgresqlSecret)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region OauthProxy Service
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
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: oauthProxyService.Name, Namespace: oauthProxyService.Namespace}, foundOauthProxyService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", oauthProxyService.Namespace, "Service.Name", oauthProxyService.Name)
		err = operatorClient.Create(context.TODO(), oauthProxyService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region UPS Service
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
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: unifiedpushService.Name, Namespace: unifiedpushService.Namespace}, foundUnifiedpushService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", unifiedpushService.Namespace, "Service.Name", unifiedpushService.Name)
		err = operatorClient.Create(context.TODO(), unifiedpushService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region OauthProxy Route
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
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: oauthProxyRoute.Name, Namespace: oauthProxyRoute.Namespace}, foundOauthProxyRoute)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Route", "Route.Namespace", oauthProxyRoute.Namespace, "Route.Name", oauthProxyRoute.Name)
		err = operatorClient.Create(context.TODO(), oauthProxyRoute)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	//#endregion

	//#region UPS DeploymentConfig
	unifiedpushDeploymentConfig, err := newUnifiedPushServerDeploymentConfig(instance)

	if err := controllerutil.SetControllerReference(instance, unifiedpushDeploymentConfig, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DeploymentConfig already exists
	foundUnifiedpushDeploymentConfig := &openshiftappsv1.DeploymentConfig{}
	err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: unifiedpushDeploymentConfig.Name, Namespace: unifiedpushDeploymentConfig.Namespace}, foundUnifiedpushDeploymentConfig)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new DeploymentConfig", "DeploymentConfig.Namespace", unifiedpushDeploymentConfig.Namespace, "DeploymentConfig.Name", unifiedpushDeploymentConfig.Name)
		err = operatorClient.Create(context.TODO(), unifiedpushDeploymentConfig)
		if err != nil {
			return reconcile.Result{}, err
		}

		// DeploymentConfig created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		unifiedPushResourceRequirements := getUnifiedPushResourceRequirements(instance)
		oauthProxyResourceRequirements := getOauthProxyResourceRequirements(instance)

		containers := foundUnifiedpushDeploymentConfig.Spec.Template.Spec.Containers
		for i := range containers {
			if containers[i].Name == cfg.UPSContainerName {
				if reflect.DeepEqual(containers[i].Resources, unifiedPushResourceRequirements) == false {
					reqLogger.Info("UnifiedPush container resource requirements are different than in the UnifiedPushServer spec or the operator defaults", "DeploymentConfig.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "DeploymentConfig.Name", foundUnifiedpushDeploymentConfig.Name, "Found resource requirements", containers[i].Resources, "Spec resource requirements", unifiedPushResourceRequirements)

					containers[i].Resources = unifiedPushResourceRequirements

					// enqueue
					err = operatorClient.Update(context.TODO(), foundUnifiedpushDeploymentConfig)
					if err != nil {
						reqLogger.Error(err, "Failed to update DeploymentConfig", "DeploymentConfig.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "DeploymentConfig.Name", foundUnifiedpushDeploymentConfig.Name)
						return reconcile.Result{}, err
					}
					return reconcile.Result{Requeue: true}, nil
				}
			} else if containers[i].Name == cfg.OauthProxyContainerName {
				if reflect.DeepEqual(containers[i].Resources, oauthProxyResourceRequirements) == false {
					reqLogger.Info("OauthProxy container resource requirements are different than in the UnifiedPushServer spec or the operator defaults", "DeploymentConfig.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "DeploymentConfig.Name", foundUnifiedpushDeploymentConfig.Name, "Found resource requirements", containers[i].Resources, "Spec resource requirements", oauthProxyResourceRequirements)

					containers[i].Resources = oauthProxyResourceRequirements

					// enqueue
					err = operatorClient.Update(context.TODO(), foundUnifiedpushDeploymentConfig)
					if err != nil {
						reqLogger.Error(err, "Failed to update DeploymentConfig", "DeploymentConfig.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "DeploymentConfig.Name", foundUnifiedpushDeploymentConfig.Name)
						return reconcile.Result{}, err
					}
					return reconcile.Result{Requeue: true}, nil
				}
			}
		}

		desiredUnifiedPushImage := constants.UPSImage
		desiredProxyImage := constants.OauthProxyImage

		unifiedPushContainerSpec := findContainerSpec(foundUnifiedpushDeploymentConfig, cfg.UPSContainerName)
		if unifiedPushContainerSpec == nil {
			reqLogger.Info("Skipping image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "Deployment.Name", foundUnifiedpushDeploymentConfig.Name, "ContainerSpec", cfg.UPSContainerName)
		} else if unifiedPushContainerSpec.Image != desiredUnifiedPushImage {
			reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "Deployment.Name", foundUnifiedpushDeploymentConfig.Name, "ContainerSpec", cfg.UPSContainerName, "ExistingImage", unifiedPushContainerSpec.Image, "DesiredImage", desiredUnifiedPushImage)

			// update
			updateContainerSpecImage(foundUnifiedpushDeploymentConfig, cfg.UPSContainerName, desiredUnifiedPushImage)

			// enqueue
			err = operatorClient.Update(context.TODO(), foundUnifiedpushDeploymentConfig)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "Deployment.Name", foundUnifiedpushDeploymentConfig.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}

		proxyContainerSpec := findContainerSpec(foundUnifiedpushDeploymentConfig, cfg.OauthProxyContainerName)
		if proxyContainerSpec == nil {
			reqLogger.Info("Skipping image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "Deployment.Name", foundUnifiedpushDeploymentConfig.Name, "ContainerSpec", cfg.OauthProxyContainerName)
		} else if proxyContainerSpec.Image != desiredProxyImage {
			reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "Deployment.Name", foundUnifiedpushDeploymentConfig.Name, "ContainerSpec", cfg.OauthProxyContainerName, "ExistingImage", proxyContainerSpec.Image, "DesiredImage", desiredProxyImage)

			// update
			updateContainerSpecImage(foundUnifiedpushDeploymentConfig, cfg.OauthProxyContainerName, desiredProxyImage)

			// enqueue
			err = operatorClient.Update(context.TODO(), foundUnifiedpushDeploymentConfig)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundUnifiedpushDeploymentConfig.Namespace, "Deployment.Name", foundUnifiedpushDeploymentConfig.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	}
	//#endregion

	//#region Backups
	if len(instance.Spec.Backups) > 0 {
		backupjobSA := &corev1.ServiceAccount{}
		err = operatorClient.Get(context.TODO(), types.NamespacedName{Name: "backupjob", Namespace: instance.Namespace}, backupjobSA)
		if err != nil {
			reqLogger.Error(err, "A 'backupjob' ServiceAccount is required for the requested backup CronJob(s). Will check again in 10 seconds")
			return reconcile.Result{RequeueAfter: time.Second * 10}, nil
		}
	}

	existingCronJobs := &batchv1beta1.CronJobList{}
	opts = client.InNamespace(instance.Namespace).MatchingLabels(labels(instance, "backup"))
	err = operatorClient.List(context.TODO(), opts, existingCronJobs)
	if err != nil {
		return reconcile.Result{}, err
	}

	desiredCronJobs, err := backups(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, desiredCronJob := range desiredCronJobs {
		if err := controllerutil.SetControllerReference(instance, &desiredCronJob, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		if exists := containsCronJob(existingCronJobs.Items, &desiredCronJob); exists {
			err = operatorClient.Update(context.TODO(), &desiredCronJob)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			reqLogger.Info("Creating a new CronJob", "CronJob.Namespace", desiredCronJob.Namespace, "CronJob.Name", desiredCronJob.Name)
			err = operatorClient.Create(context.TODO(), &desiredCronJob)
			if err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	}

	for _, existingCronJob := range existingCronJobs.Items {
		desired := containsCronJob(desiredCronJobs, &existingCronJob)
		if !desired {
			reqLogger.Info("Deleting backup CronJob since it was removed from CR", "CronJob.Namespace", existingCronJob.Namespace, "CronJob.Name", existingCronJob.Name)
			err = operatorClient.Delete(context.TODO(), &existingCronJob)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}
	//#endregion

	if foundUnifiedpushDeploymentConfig.Status.ReadyReplicas > 0 && instance.Status.Phase != pushv1alpha1.PhaseComplete {
		instance.Status.Phase = pushv1alpha1.PhaseComplete
		operatorClient.Status().Update(context.TODO(), instance)
	}

	// Resources already exist - don't requeue
	reqLogger.Info("Skip reconcile: Resources already exist")
	return reconcile.Result{}, nil
}

func getPostgresResourceRequirements(instance *pushv1alpha1.UnifiedPushServer) corev1.ResourceRequirements {
	postgresDefaultResourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"memory": resource.MustParse(cfg.PostgresMemoryLimit),
			"cpu":    resource.MustParse(cfg.PostgresCpuLimit),
		},
		Requests: corev1.ResourceList{
			"memory": resource.MustParse(cfg.PostgresMemoryRequest),
			"cpu":    resource.MustParse(cfg.PostgresCpuRequest),
		},
	}

	return applyDefaultsToResourceRequirements(instance.Spec.PostgresResourceRequirements, postgresDefaultResourceRequirements)
}

func getUnifiedPushResourceRequirements(instance *pushv1alpha1.UnifiedPushServer) corev1.ResourceRequirements {
	unifiedPushDefaultResourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"memory": resource.MustParse(cfg.UPSMemoryLimit),
			"cpu":    resource.MustParse(cfg.UPSCpuLimit),
		},
		Requests: corev1.ResourceList{
			"memory": resource.MustParse(cfg.UPSMemoryRequest),
			"cpu":    resource.MustParse(cfg.UPSCpuRequest),
		},
	}

	return applyDefaultsToResourceRequirements(instance.Spec.UnifiedPushResourceRequirements, unifiedPushDefaultResourceRequirements)
}

func getOauthProxyResourceRequirements(instance *pushv1alpha1.UnifiedPushServer) corev1.ResourceRequirements {
	oauthProxyDefaultResourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"memory": resource.MustParse(cfg.OauthMemoryLimit),
			"cpu":    resource.MustParse(cfg.OauthCpuLimit),
		},
		Requests: corev1.ResourceList{
			"memory": resource.MustParse(cfg.OauthMemoryRequest),
			"cpu":    resource.MustParse(cfg.OauthCpuRequest),
		},
	}

	return applyDefaultsToResourceRequirements(instance.Spec.OAuthResourceRequirements, oauthProxyDefaultResourceRequirements)
}

func getPostgresPVCSize(instance *pushv1alpha1.UnifiedPushServer) string {
	requiredPostgresPVCSize := instance.Spec.PostgresPVCSize
	if requiredPostgresPVCSize == "" {
		requiredPostgresPVCSize = cfg.PostgresPVCSize
	}
	return requiredPostgresPVCSize
}

func applyDefaultsToResourceRequirements(reqs corev1.ResourceRequirements, defs corev1.ResourceRequirements) corev1.ResourceRequirements {
	if reqs.Requests == nil {
		reqs.Requests = make(map[corev1.ResourceName]resource.Quantity)
	}

	if reqs.Limits == nil {
		reqs.Limits = make(map[corev1.ResourceName]resource.Quantity)
	}

	for k, v := range defs.Requests {
		if _, ok := reqs.Requests[k]; !ok {
			reqs.Requests[k] = v
		}
	}

	for k, v := range defs.Limits {
		if _, ok := reqs.Limits[k]; !ok {
			reqs.Limits[k] = v
		}
	}

	return reqs
}

func containsCronJob(cronJobs []batchv1beta1.CronJob, candidate *batchv1beta1.CronJob) bool {
	for _, cronJob := range cronJobs {
		if candidate.Name == cronJob.Name {
			return true
		}
	}
	return false
}
