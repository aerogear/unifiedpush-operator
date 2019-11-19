package unifiedpushserver

import (
	"context"

	"os"
	"reflect"
	"time"

	"github.com/aerogear/unifiedpush-operator/pkg/constants"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	"github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	"github.com/aerogear/unifiedpush-operator/pkg/config"
	"github.com/aerogear/unifiedpush-operator/pkg/nspredicate"

	enmassev1beta "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1beta1"
	messaginguserv1beta "github.com/enmasseproject/enmasse/pkg/apis/user/v1beta1"

	openshiftappsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
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

const (
	controllerName    = "unifiedpushserver-controller"
	requeueDelay      = 30 * time.Second
	requeueErrorDelay = 5 * time.Second
)

var (
	cfg = config.New()
	log = logf.Log.WithName(controllerName)
)

// Add creates a new UnifiedPushServer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Error(err, "Failed to get clientset")
		os.Exit(1)
	}
	return &ReconcileUnifiedPushServer{
		client:            mgr.GetClient(),
		scheme:            mgr.GetScheme(),
		config:            mgr.GetConfig(),
		apiVersionChecker: getApiVersionChecker(clientset),
		recorder:          mgr.GetRecorder(controllerName),
		readyStatus:       true,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
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

	// Watch for changes to secondary resource ImageStream and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &imagev1.ImageStream{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
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

	// Watch for changes to secondary resource ServiceMonitor and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &monitoringv1.ServiceMonitor{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	// If the problem is just a missing kind, don't worry about it
	if _, isNoKindMatchError := err.(*meta.NoKindMatchError); err != nil && !isNoKindMatchError {
		return err
	}

	// Watch for changes to secondary resource PrometheusRule and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &monitoringv1.PrometheusRule{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pushv1alpha1.UnifiedPushServer{},
	})
	// If the problem is just a missing kind, don't worry about it
	if _, isNoKindMatchError := err.(*meta.NoKindMatchError); err != nil && !isNoKindMatchError {
		return err
	}

	// Watch for changes to secondary resource GrafanaDashboard and requeue the owner UnifiedPushServer
	err = c.Watch(&source.Kind{Type: &integreatlyv1alpha1.GrafanaDashboard{}}, &handler.EnqueueRequestForOwner{
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
	client            client.Client
	scheme            *runtime.Scheme
	config            *rest.Config
	apiVersionChecker *apiVersionChecker
	recorder          record.EventRecorder
	readyStatus       bool
}

// Reconcile reads the state of the cluster for a UnifiedPushServer object and makes changes based on the state read
// and what is in the UnifiedPushServer.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileUnifiedPushServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling UnifiedPushServer")

	secondaryResources := resources{}

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
		return r.manageError(instance, err)
	}

	// look for other unifiedPush resources and don't provision a new one if there is another one with Phase=Complete
	existingInstances := &pushv1alpha1.UnifiedPushServerList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UnifiedPushServer",
			APIVersion: "push.aerogear.org/v1alpha1",
		},
	}
	opts := &client.ListOptions{Namespace: instance.Namespace}
	err = r.client.List(context.TODO(), opts, existingInstances)
	if err != nil {
		reqLogger.Error(err, "Failed to list UnifiedPush resources", "UnifiedPush.Namespace", instance.Namespace)
		return r.manageError(instance, err)
	} else if len(existingInstances.Items) > 1 { // check if > 1 since there's the current one already in that list.
		for _, existingInstance := range existingInstances.Items {
			if existingInstance.Name == instance.Name {
				continue
			}
			if existingInstance.Status.Phase == pushv1alpha1.PhaseInitializing || existingInstance.Status.Phase == pushv1alpha1.PhaseReconciling || existingInstance.Status.Phase == pushv1alpha1.PhaseFailing {
				reqLogger.Info("There is already a UnifiedPush resource in Complete phase. Doing nothing for this CR.", "UnifiedPush.Namespace", instance.Namespace, "UnifiedPush.Name", instance.Name)
				return reconcile.Result{}, nil
			}
		}
	}

	if instance.Status.Phase == pushv1alpha1.PhaseEmpty {
		instance.Status.Phase = pushv1alpha1.PhaseInitializing
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update UnifiedPush resource status phase", "UnifiedPush.Namespace", instance.Namespace, "UnifiedPush.Name", instance.Name)
			return r.manageError(instance, err)
		}
	}

	//#region AMQ resource reconcile
	if instance.Spec.UseMessageBroker {
		//#region create addressSpace
		addressSpace := newAddressSpace(instance)

		// Set UnifiedPushServer instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, addressSpace, r.scheme); err != nil {
			return r.manageError(instance, err)
		}

		foundAddressSpace := &enmassev1beta.AddressSpace{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: addressSpace.Name, Namespace: addressSpace.Namespace}, foundAddressSpace)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Address Space", "AddressSpace.Namespace", addressSpace.Namespace, "AddressSpace.Name", addressSpace.Name)
			err = r.client.Create(context.TODO(), addressSpace)
			if err != nil {
				return r.manageError(instance, err)
			}
			reqLogger.Info("Requeuing, AddressSpace not ready.", "AddressSpace.Namespace", addressSpace.Namespace, "AddressSpace.Name", addressSpace.Name)
			return reconcile.Result{RequeueAfter: time.Second * 10}, nil
		} else if err != nil {
			return r.manageError(instance, err)
		} else if !foundAddressSpace.Status.IsReady {
			reqLogger.Info("Requeuing, AddressSpace not ready.", "AddressSpace.Namespace", foundAddressSpace.Namespace, "AddressSpace.Name", foundAddressSpace.Name)
			return reconcile.Result{RequeueAfter: time.Second * 10}, nil
		} else {
			reqLogger.Info("Found AddressSpace for UPS")
		}

		secondaryResources.add("AddressSpace", addressSpace.Name)
		//#endregion

		//#region check that user exists
		user, err := newMessagingUser(instance)
		if err != nil {
			return r.manageError(instance, err)
		}

		// Set UnifiedPushServer instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, user, r.scheme); err != nil {
			return r.manageError(instance, err)
		}

		foundUser := &messaginguserv1beta.MessagingUser{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: user.Name, Namespace: user.Namespace}, foundUser)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new MessagingUser", "MessagingUser.Namespace", user.Namespace, "MessagingUser.Name", user.Name)
			err = r.client.Create(context.TODO(), user)
			if err != nil {
				return r.manageError(instance, err)
			}

		} else if err != nil {
			return r.manageError(instance, err)
		}
		secondaryResources.add("MessagingUser", user.Name)
		//#endregion

		//#region create secret for user password and artemis url
		for _, status := range foundAddressSpace.Status.EndpointStatus {
			if status.Name == "messaging" { //"messaging" is a key from enmasse.
				addressSpaceURL := status.ServiceHost
				password := string(user.Spec.Authentication.Password)
				secret := newAMQSecret(instance, password, addressSpaceURL)
				foundSecret := &corev1.Secret{}
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, foundSecret)
				if err != nil && errors.IsNotFound(err) {
					reqLogger.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
					err = r.client.Create(context.TODO(), secret)
					if err != nil {
						return r.manageError(instance, err)
					}
				} else if err != nil {
					return r.manageError(instance, err)
				}
				secondaryResources.add("Secret", secret.Name)
				break
			}
		}
		//#endregion

		//#region queues
		queues := []string{"APNsPushMessageQueue", "APNsTokenBatchQueue", "GCMPushMessageQueue", "GCMTokenBatchQueue", "WNSPushMessageQueue", "WNSTokenBatchQueue", "WebPushMessageQueue", "WebTokenBatchQueue", "MetricsQueue", "TriggerMetricCollectionQueue", "TriggerVariantMetricCollectionQueue", "BatchLoadedQueue", "AllBatchesLoadedQueue", "FreeServiceSlotQueue"}
		requeueCreate := false
		for _, address := range queues {
			queue := newQueue(instance, address)
			foundQueue := &enmassev1beta.Address{}
			// Set UnifiedPushServer instance as the owner and controller
			if err := controllerutil.SetControllerReference(instance, queue, r.scheme); err != nil {
				return r.manageError(instance, err)
			}

			err = r.client.Get(context.TODO(), types.NamespacedName{Name: queue.Name, Namespace: queue.Namespace}, foundQueue)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Queue", "Queue.Namespace", queue.Namespace, "Queue.Name", queue.Name)
				err = r.client.Create(context.TODO(), queue)
				if err != nil {
					return r.manageError(instance, err)
				}
				requeueCreate = true
			} else if err != nil {
				reqLogger.Info("Queue Error")
				return r.manageError(instance, err)
			} else if !foundQueue.Status.IsReady {
				reqLogger.Info("Queue Not ready", "Queue.Name", foundQueue.Name)
				requeueCreate = true
			}
			secondaryResources.add("Address", queue.Name)
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
				return r.manageError(instance, err)
			}

			err = r.client.Get(context.TODO(), types.NamespacedName{Name: topic.Name, Namespace: topic.Namespace}, foundTopic)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Topic", "Topic.Namespace", topic.Namespace, "Topic.Name", topic.Name)
				err = r.client.Create(context.TODO(), topic)
				if err != nil {
					return r.manageError(instance, err)
				}
				requeueCreate = true
			} else if err != nil {
				return r.manageError(instance, err)
			}
			secondaryResources.add("Address", topic.Name)
		}

		if requeueCreate {
			reqLogger.Info("Requeueing while topics are created")
			return reconcile.Result{RequeueAfter: time.Second * 5}, nil
		}
		//#endregion

		reqLogger.Info("Found All queues and topics for UPS")

	}
	//#endregion

	//#region MIGRATION from old resources to new ones

	// TODO: This migration block should be removed after a major release!
	// TODO: in UPS operator version <0.3.0, we were using DCs and ImageStreams.
	// TODO: in 0.3.0, we introduced this code block to migrate from old resources to new ones.
	// TODO: in a future version we should get rid of this migration block, as well as
	// TODO: unneeded permissions to access these old resources.
	dcResourceExists, err := r.apiVersionChecker.check("apps.openshift.io/v1")
	if err != nil {
		reqLogger.Error(err, "Unable to check if a OpenShift's apps.openshift.io/v1 api version is available.")
		return r.manageError(instance, err)
	}

	imageStreamResourceExists, err := r.apiVersionChecker.check("image.openshift.io/v1")
	if err != nil {
		reqLogger.Error(err, "Unable to check if a OpenShift's image.openshift.io/v1 api version is available.")
		return r.manageError(instance, err)
	}

	if dcResourceExists {
		//#region DELETE UnifiedPush Server DeploymentConfig as we moved to Kube Deployments now
		upsDeploymentConfigObjectMeta := metav1.ObjectMeta{
			Namespace: instance.Namespace,
			Name:      instance.Name, // this is the name of the DeploymentConfig we were using
		}
		foundUpsDeploymentConfig := &openshiftappsv1.DeploymentConfig{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: upsDeploymentConfigObjectMeta.Name, Namespace: upsDeploymentConfigObjectMeta.Namespace}, foundUpsDeploymentConfig)
		if err != nil && !errors.IsNotFound(err) {
			// if there is another error than the DC not being found
			reqLogger.Error(err, "Unable to check if a DeploymentConfig exists for UnifiedPush Server.", "DeploymentConfig.Namespace", foundUpsDeploymentConfig.Namespace, "DeploymentConfig.Name", foundUpsDeploymentConfig.Name)
			return r.manageError(instance, err)
		} else if err == nil {
			reqLogger.Info("Found a DeploymentConfig for UnifiedPush Server. Deleting it.", "DeploymentConfig.Namespace", foundUpsDeploymentConfig.Namespace, "DeploymentConfig.Name", foundUpsDeploymentConfig.Name)
			err = r.client.Delete(context.TODO(), foundUpsDeploymentConfig)
			if err != nil {
				reqLogger.Error(err, "Unable to delete the DeploymentConfig for UnifiedPush Server.", "DeploymentConfig.Namespace", foundUpsDeploymentConfig.Namespace, "DeploymentConfig.Name", foundUpsDeploymentConfig.Name)
				return r.manageError(instance, err)
			}
		}
		//#endregion

		//#region DELETE Postgres DeploymentConfig as we moved to Kube Deployments now
		postgresDeploymentConfigObjectMeta := objectMeta(instance, "postgresql")
		foundPostgresqlDeploymentConfig := &openshiftappsv1.DeploymentConfig{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: postgresDeploymentConfigObjectMeta.Name, Namespace: postgresDeploymentConfigObjectMeta.Namespace}, foundPostgresqlDeploymentConfig)
		if err != nil && !errors.IsNotFound(err) {
			// if there is another error than the DC not being found
			reqLogger.Error(err, "Unable to check if a DeploymentConfig exists for Postgres.", "DeploymentConfig.Namespace", foundPostgresqlDeploymentConfig.Namespace, "DeploymentConfig.Name", foundPostgresqlDeploymentConfig.Name)
			return r.manageError(instance, err)
		} else if err == nil {
			reqLogger.Info("Found a DeploymentConfig for Postgres. Deleting it.", "DeploymentConfig.Namespace", foundPostgresqlDeploymentConfig.Namespace, "DeploymentConfig.Name", foundPostgresqlDeploymentConfig.Name)
			err = r.client.Delete(context.TODO(), foundPostgresqlDeploymentConfig)
			if err != nil {
				reqLogger.Error(err, "Unable to delete the DeploymentConfig for Postgres.", "DeploymentConfig.Namespace", foundPostgresqlDeploymentConfig.Namespace, "DeploymentConfig.Name", foundPostgresqlDeploymentConfig.Name)
				return r.manageError(instance, err)
			}
		}
		//#endregion
	}

	if imageStreamResourceExists {
		//#region DELETE OAuth Proxy ImageStream as we moved to using static image references
		oauthProxyImageStreamObjectMeta := metav1.ObjectMeta{
			Namespace: instance.Namespace,
			Name:      "ups-oauth-proxy-imagestream", // this is the name of the image stream we were using
		}
		foundOAuthProxyImageStream := &imagev1.ImageStream{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: oauthProxyImageStreamObjectMeta.Name, Namespace: oauthProxyImageStreamObjectMeta.Namespace}, foundOAuthProxyImageStream)
		if err != nil && !errors.IsNotFound(err) {
			// if there is another error than the DC not being found
			reqLogger.Error(err, "Unable to check if a ImageStream exists for OAuth Proxy.", "ImageStream.Namespace", foundOAuthProxyImageStream.Namespace, "ImageStream.Name", foundOAuthProxyImageStream.Name)
			// don't do anything.
			// we simply log this, and it should be ok to have some leftover ImageStreams
		} else if err == nil {
			reqLogger.Info("Found a ImageStream for OAuth Proxy. Deleting it.", "ImageStream.Namespace", foundOAuthProxyImageStream.Namespace, "ImageStream.Name", foundOAuthProxyImageStream.Name)
			err = r.client.Delete(context.TODO(), foundOAuthProxyImageStream)
			if err != nil {
				reqLogger.Error(err, "Unable to delete ImageStream. Skipping it.", "ImageStream.Namespace", foundOAuthProxyImageStream.Namespace, "ImageStream.Name", foundOAuthProxyImageStream.Name)
				// don't do anything.
				// we simply log this, and it should be ok to have some leftover ImageStreams
			}
		}
		//#endregion

		//#region DELETE UnifiedPush Server ImageStream as we moved to using static image references
		upsImageStreamObjectMeta := metav1.ObjectMeta{
			Namespace: instance.Namespace,
			Name:      "ups-imagestream", // this is the name of the image stream we were using
		}
		foundUpsImageStream := &imagev1.ImageStream{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: upsImageStreamObjectMeta.Name, Namespace: upsImageStreamObjectMeta.Namespace}, foundUpsImageStream)
		if err != nil && !errors.IsNotFound(err) {
			// if there is another error than the ImageStream not being found
			reqLogger.Error(err, "Unable to check if an ImageStream exists for UnifiedPush Server.", "ImageStream.Namespace", foundUpsImageStream.Namespace, "ImageStream.Name", foundUpsImageStream.Name)
			// don't do anything.
			// we simply log this, and it should be ok to have some leftover ImageStreams
		} else if err == nil {
			reqLogger.Info("Found an ImageStream for UnifiedPush Server. Deleting it.", "ImageStream.Namespace", foundUpsImageStream.Namespace, "ImageStream.Name", foundUpsImageStream.Name)
			err = r.client.Delete(context.TODO(), foundUpsImageStream)
			if err != nil {
				reqLogger.Error(err, "Unable to delete ImageStream. Skipping it.", "ImageStream.Namespace", foundUpsImageStream.Namespace, "ImageStream.Name", foundUpsImageStream.Name)
				// don't do anything.
				// we simply log this, and it should be ok to have some leftover ImageStreams
			}
		}
		//#endregion
	}

	//#endregion

	if !instance.Spec.ExternalDB {

		//#region Postgres PVC
		persistentVolumeClaim, err := newPostgresqlPersistentVolumeClaim(instance)
		if err != nil {
			return r.manageError(instance, err)
		}

		// Set UnifiedPushServer instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, persistentVolumeClaim, r.scheme); err != nil {
			return r.manageError(instance, err)
		}

		// Check if this PersistentVolumeClaim already exists
		foundPersistentVolumeClaim := &corev1.PersistentVolumeClaim{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: persistentVolumeClaim.Name, Namespace: persistentVolumeClaim.Namespace}, foundPersistentVolumeClaim)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", persistentVolumeClaim.Namespace, "PersistentVolumeClaim.Name", persistentVolumeClaim.Name)
			err = r.client.Create(context.TODO(), persistentVolumeClaim)
			if err != nil {
				return r.manageError(instance, err)
			}
		} else if err != nil {
			return r.manageError(instance, err)
		} else {
			requiredPostgresPVCSize := getPostgresPVCSize(instance)

			foundPVCSize := foundPersistentVolumeClaim.Spec.Resources.Requests[corev1.ResourceStorage]
			if foundPVCSize.String() != requiredPostgresPVCSize {
				reqLogger.Info("Request size of PersistentVolumeClaim is different than in the UnifiedPushServer spec or the operator defaults", "PersistentVolumeClaim.Namespace", persistentVolumeClaim.Namespace, "PersistentVolumeClaim.Name", persistentVolumeClaim.Name, "Found size", foundPVCSize.String(), "Spec size", requiredPostgresPVCSize)

				foundPersistentVolumeClaim.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(requiredPostgresPVCSize)

				// enqueue
				err = r.client.Update(context.TODO(), foundPersistentVolumeClaim)
				if err != nil {
					reqLogger.Error(err, "Failed to update PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", foundPersistentVolumeClaim.Namespace, "PersistentVolumeClaim.Name", foundPersistentVolumeClaim.Name)
					return r.manageError(instance, err)
				}
				return reconcile.Result{Requeue: true}, nil
			}
		}
		secondaryResources.add("PersistentVolumeClaim", persistentVolumeClaim.Name)
		//#endregion

		//#region Postgres Deployment
		postgresqlDeployment, err := newPostgresqlDeployment(instance)
		if err != nil {
			return r.manageError(instance, err)
		}

		// Set UnifiedPushServer instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, postgresqlDeployment, r.scheme); err != nil {
			return r.manageError(instance, err)
		}

		// Check if this Deployment already exists
		foundPostgresqlDeployment := &appsv1.Deployment{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: postgresqlDeployment.Name, Namespace: postgresqlDeployment.Namespace}, foundPostgresqlDeployment)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", postgresqlDeployment.Namespace, "Deployment.Name", postgresqlDeployment.Name)
			err = r.client.Create(context.TODO(), postgresqlDeployment)
			if err != nil {
				return r.manageError(instance, err)
			}
		} else if err != nil {
			return r.manageError(instance, err)
		} else {
			postgresResourceRequirements := getPostgresResourceRequirements(instance)

			containers := foundPostgresqlDeployment.Spec.Template.Spec.Containers
			for i := range containers {
				if containers[i].Name == cfg.PostgresContainerName {
					if reflect.DeepEqual(containers[i].Resources, postgresResourceRequirements) == false {
						reqLogger.Info("Postgres container resource requirements are different than in the UnifiedPushServer spec or the operator defaults", "Deployment.Namespace", foundPostgresqlDeployment.Namespace, "Deployment.Name", foundPostgresqlDeployment.Name, "Found resource requirements", containers[i].Resources, "Spec resource requirements", postgresResourceRequirements)

						containers[i].Resources = postgresResourceRequirements

						// enqueue
						err = r.client.Update(context.TODO(), foundPostgresqlDeployment)
						if err != nil {
							reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundPostgresqlDeployment.Namespace, "Deployment.Name", foundPostgresqlDeployment.Name)
							return r.manageError(instance, err)
						}
						return reconcile.Result{Requeue: true}, nil
					}
				}
			}

			desiredImage := constants.PostgresImage

			containerSpec := findContainerSpec(foundPostgresqlDeployment, cfg.PostgresContainerName)
			if containerSpec == nil {
				reqLogger.Info("Unable to do image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundPostgresqlDeployment.Namespace, "Deployment.Name", foundPostgresqlDeployment.Name, "ContainerSpec", cfg.PostgresContainerName)
				return reconcile.Result{Requeue: true}, nil
			} else if containerSpec.Image != desiredImage {
				reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundPostgresqlDeployment.Namespace, "Deployment.Name", foundPostgresqlDeployment.Name, "ContainerSpec", cfg.PostgresContainerName, "ExistingImage", containerSpec.Image, "DesiredImage", desiredImage)

				// update
				updateContainerSpecImage(foundPostgresqlDeployment, cfg.PostgresContainerName, desiredImage)

				// enqueue
				err = r.client.Update(context.TODO(), foundPostgresqlDeployment)
				if err != nil {
					reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundPostgresqlDeployment.Namespace, "Deployment.Name", foundPostgresqlDeployment.Name)
					return r.manageError(instance, err)
				}
				return reconcile.Result{Requeue: true}, nil
			}

			// Set ready status
			deploymentReady, err := isDeploymentReady(foundPostgresqlDeployment)
			if err != nil {
				return r.manageError(instance, err)
			}
			r.readyStatus = r.readyStatus && deploymentReady
		}
		secondaryResources.add("Deployment", postgresqlDeployment.Name)
		//#endregion

		//#region Postgres Service
		postgresqlService, err := newPostgresqlService(instance)
		if err != nil {
			return r.manageError(instance, err)
		}

		// Set UnifiedPushServer instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, postgresqlService, r.scheme); err != nil {
			return r.manageError(instance, err)
		}

		// Check if this Service already exists
		foundPostgresqlService := &corev1.Service{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: postgresqlService.Name, Namespace: postgresqlService.Namespace}, foundPostgresqlService)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Service", "Service.Namespace", postgresqlService.Namespace, "Service.Name", postgresqlService.Name)
			err = r.client.Create(context.TODO(), postgresqlService)
			if err != nil {
				return r.manageError(instance, err)
			}
		} else if err != nil {
			return r.manageError(instance, err)
		}

		secondaryResources.add("Service", postgresqlService.Name)
		//#endregion
	}

	//#region ServiceAccount
	serviceAccount, err := newUnifiedPushServiceAccount(instance)

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, serviceAccount, r.scheme); err != nil {
		return r.manageError(instance, err)
	}

	// Check if this ServiceAccount already exists
	foundServiceAccount := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
		err = r.client.Create(context.TODO(), serviceAccount)
		if err != nil {
			return r.manageError(instance, err)
		}
	} else if err != nil {
		return r.manageError(instance, err)
	}
	secondaryResources.add("ServiceAccount", serviceAccount.Name)
	//#endregion

	//#region Postgres Secret
	postgresqlSecret, err := newPostgresqlSecret(instance)
	if err != nil {
		return r.manageError(instance, err)
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, postgresqlSecret, r.scheme); err != nil {
		return r.manageError(instance, err)
	}

	// Check if this Secret already exists
	foundPostgresqlSecret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: postgresqlSecret.Name, Namespace: postgresqlSecret.Namespace}, foundPostgresqlSecret)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", postgresqlSecret.Namespace, "Secret.Name", postgresqlSecret.Name)
		err = r.client.Create(context.TODO(), postgresqlSecret)
		if err != nil {
			return r.manageError(instance, err)
		}
	} else if err != nil {
		return r.manageError(instance, err)
	}
	secondaryResources.add("Secret", postgresqlSecret.Name)
	//#endregion

	//#region OauthProxy Service
	oauthProxyService, err := newOauthProxyService(instance)
	if err != nil {
		return r.manageError(instance, err)
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, oauthProxyService, r.scheme); err != nil {
		return r.manageError(instance, err)
	}

	// Check if this Service already exists
	foundOauthProxyService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: oauthProxyService.Name, Namespace: oauthProxyService.Namespace}, foundOauthProxyService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", oauthProxyService.Namespace, "Service.Name", oauthProxyService.Name)
		err = r.client.Create(context.TODO(), oauthProxyService)
		if err != nil {
			return r.manageError(instance, err)
		}
	} else if err != nil {
		return r.manageError(instance, err)
	}
	secondaryResources.add("Service", oauthProxyService.Name)
	//#endregion

	//#region UPS Service
	unifiedpushService, err := newUnifiedPushServerService(instance)
	if err != nil {
		return r.manageError(instance, err)
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, unifiedpushService, r.scheme); err != nil {
		return r.manageError(instance, err)
	}

	// Check if this Service already exists
	foundUnifiedpushService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: unifiedpushService.Name, Namespace: unifiedpushService.Namespace}, foundUnifiedpushService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", unifiedpushService.Namespace, "Service.Name", unifiedpushService.Name)
		err = r.client.Create(context.TODO(), unifiedpushService)
		if err != nil {
			return r.manageError(instance, err)
		}
	} else if err != nil {
		return r.manageError(instance, err)
	}
	secondaryResources.add("Service", unifiedpushService.Name)
	//#endregion

	//#region OauthProxy Route
	oauthProxyRoute, err := newOauthProxyRoute(instance)
	if err != nil {
		return r.manageError(instance, err)
	}

	// Set UnifiedPushServer instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, oauthProxyRoute, r.scheme); err != nil {
		return r.manageError(instance, err)
	}

	// Check if this Route already exists
	foundOauthProxyRoute := &routev1.Route{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: oauthProxyRoute.Name, Namespace: oauthProxyRoute.Namespace}, foundOauthProxyRoute)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Route", "Route.Namespace", oauthProxyRoute.Namespace, "Route.Name", oauthProxyRoute.Name)
		err = r.client.Create(context.TODO(), oauthProxyRoute)
		if err != nil {
			return r.manageError(instance, err)
		}
	} else if err != nil {
		return r.manageError(instance, err)
	}

	r.readyStatus = r.readyStatus && isRouteReady(foundOauthProxyRoute)
	secondaryResources.add("Route", oauthProxyRoute.Name)
	//#endregion

	//#region UPS Deployment
	unifiedpushDeployment, err := newUnifiedPushServerDeployment(instance)

	if err := controllerutil.SetControllerReference(instance, unifiedpushDeployment, r.scheme); err != nil {
		return r.manageError(instance, err)
	}

	// Check if this Deployment already exists
	foundUnifiedpushDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: unifiedpushDeployment.Name, Namespace: unifiedpushDeployment.Namespace}, foundUnifiedpushDeployment)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", unifiedpushDeployment.Namespace, "Deployment.Name", unifiedpushDeployment.Name)
		err = r.client.Create(context.TODO(), unifiedpushDeployment)
		if err != nil {
			return r.manageError(instance, err)
		}

		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return r.manageError(instance, err)
	} else {
		unifiedPushResourceRequirements := getUnifiedPushResourceRequirements(instance)
		oauthProxyResourceRequirements := getOauthProxyResourceRequirements(instance)

		containers := foundUnifiedpushDeployment.Spec.Template.Spec.Containers
		for i := range containers {
			if containers[i].Name == cfg.UPSContainerName {
				if reflect.DeepEqual(containers[i].Resources, unifiedPushResourceRequirements) == false {
					reqLogger.Info("UnifiedPush container resource requirements are different than in the UnifiedPushServer spec or the operator defaults", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "Found resource requirements", containers[i].Resources, "Spec resource requirements", unifiedPushResourceRequirements)

					containers[i].Resources = unifiedPushResourceRequirements

					// enqueue
					err = r.client.Update(context.TODO(), foundUnifiedpushDeployment)
					if err != nil {
						reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name)
						return r.manageError(instance, err)
					}
					return reconcile.Result{Requeue: true}, nil
				}
			} else if containers[i].Name == cfg.OauthProxyContainerName {
				if reflect.DeepEqual(containers[i].Resources, oauthProxyResourceRequirements) == false {
					reqLogger.Info("OauthProxy container resource requirements are different than in the UnifiedPushServer spec or the operator defaults", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "Found resource requirements", containers[i].Resources, "Spec resource requirements", oauthProxyResourceRequirements)

					containers[i].Resources = oauthProxyResourceRequirements

					// enqueue
					err = r.client.Update(context.TODO(), foundUnifiedpushDeployment)
					if err != nil {
						reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name)
						return r.manageError(instance, err)
					}
					return reconcile.Result{Requeue: true}, nil
				}
			}
		}

		desiredUnifiedPushImage := constants.UPSImage
		desiredProxyImage := constants.OauthProxyImage

		unifiedPushContainerSpec := findContainerSpec(foundUnifiedpushDeployment, cfg.UPSContainerName)
		if unifiedPushContainerSpec == nil {
			reqLogger.Info("Unable to do image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "ContainerSpec", cfg.UPSContainerName)
			return reconcile.Result{Requeue: true}, nil
		} else if unifiedPushContainerSpec.Image != desiredUnifiedPushImage {
			reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "ContainerSpec", cfg.UPSContainerName, "ExistingImage", unifiedPushContainerSpec.Image, "DesiredImage", desiredUnifiedPushImage)

			// update
			updateContainerSpecImage(foundUnifiedpushDeployment, cfg.UPSContainerName, desiredUnifiedPushImage)

			// enqueue
			err = r.client.Update(context.TODO(), foundUnifiedpushDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name)
				return r.manageError(instance, err)
			}
			return reconcile.Result{Requeue: true}, nil
		}

		proxyContainerSpec := findContainerSpec(foundUnifiedpushDeployment, cfg.OauthProxyContainerName)
		if proxyContainerSpec == nil {
			reqLogger.Info("Unable to do image reconcile: Unable to find container spec in deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "ContainerSpec", cfg.OauthProxyContainerName)
			return reconcile.Result{Requeue: true}, nil
		} else if proxyContainerSpec.Image != desiredProxyImage {
			reqLogger.Info("Container spec in deployment is using a different image. Going to update it now.", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name, "ContainerSpec", cfg.OauthProxyContainerName, "ExistingImage", proxyContainerSpec.Image, "DesiredImage", desiredProxyImage)

			// update
			updateContainerSpecImage(foundUnifiedpushDeployment, cfg.OauthProxyContainerName, desiredProxyImage)

			// enqueue
			err = r.client.Update(context.TODO(), foundUnifiedpushDeployment)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundUnifiedpushDeployment.Namespace, "Deployment.Name", foundUnifiedpushDeployment.Name)
				return r.manageError(instance, err)
			}
			return reconcile.Result{Requeue: true}, nil
		}

		// Set ready status
		deploymentReady, err := isDeploymentReady(foundUnifiedpushDeployment)
		if err != nil {
			return r.manageError(instance, err)
		}
		r.readyStatus = r.readyStatus && deploymentReady
	}
	secondaryResources.add("Deployment", unifiedpushDeployment.Name)
	//#endregion

	//#region Backups
	if len(instance.Spec.Backups) > 0 {
		backupjobSA := &corev1.ServiceAccount{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: "backupjob", Namespace: instance.Namespace}, backupjobSA)
		if err != nil {
			reqLogger.Error(err, "A 'backupjob' ServiceAccount is required for the requested backup CronJob(s). Will check again in 10 seconds")
			return reconcile.Result{RequeueAfter: time.Second * 10}, nil
		}
	}

	existingCronJobs := &batchv1beta1.CronJobList{}
	opts = client.InNamespace(instance.Namespace).MatchingLabels(labels(instance, "backup"))
	err = r.client.List(context.TODO(), opts, existingCronJobs)
	if err != nil {
		return r.manageError(instance, err)
	}

	desiredCronJobs, err := backups(instance)
	if err != nil {
		return r.manageError(instance, err)
	}

	for _, desiredCronJob := range desiredCronJobs {
		if err := controllerutil.SetControllerReference(instance, &desiredCronJob, r.scheme); err != nil {
			return r.manageError(instance, err)
		}

		if exists := containsCronJob(existingCronJobs.Items, &desiredCronJob); exists {
			err = r.client.Update(context.TODO(), &desiredCronJob)
			if err != nil {
				return r.manageError(instance, err)
			}
		} else {
			reqLogger.Info("Creating a new CronJob", "CronJob.Namespace", desiredCronJob.Namespace, "CronJob.Name", desiredCronJob.Name)
			err = r.client.Create(context.TODO(), &desiredCronJob)
			if err != nil {
				return r.manageError(instance, err)
			}
			return reconcile.Result{}, nil
		}
		secondaryResources.add("CronJob", desiredCronJob.Name)
	}

	for _, existingCronJob := range existingCronJobs.Items {
		desired := containsCronJob(desiredCronJobs, &existingCronJob)
		if !desired {
			reqLogger.Info("Deleting backup CronJob since it was removed from CR", "CronJob.Namespace", existingCronJob.Namespace, "CronJob.Name", existingCronJob.Name)
			err = r.client.Delete(context.TODO(), &existingCronJob)
			if err != nil {
				return r.manageError(instance, err)
			}
			secondaryResources.remove("CronJob", existingCronJob.Name)
		}
	}
	//#endregion

	//#region Monitoring
	//## region ServiceMonitor
	serviceMonitor := &monitoringv1.ServiceMonitor{ObjectMeta: metav1.ObjectMeta{Name: "unifiedpush", Namespace: instance.Namespace}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.client, serviceMonitor, func(ignore runtime.Object) error {
		reconcileServiceMonitor(serviceMonitor)
		// Set UnifiedPushServer instance as the owner and controller
		err := controllerutil.SetControllerReference(instance, serviceMonitor, r.scheme)
		return err
	})
	if err != nil {
		return r.manageError(instance, err)
	}
	if op != controllerutil.OperationResultNone {
		reqLogger.Info("ServiceMonitor reconciled:", "ServiceMonitor.Name", serviceMonitor.Name, "ServiceMonitor.Namespace", serviceMonitor.Namespace, "Operation", op)
	}
	//## endregion ServiceMonitor

	//## region PrometheusRule
	prometheusRule := &monitoringv1.PrometheusRule{ObjectMeta: metav1.ObjectMeta{Name: "unifiedpush", Namespace: instance.Namespace}}
	op, err = controllerutil.CreateOrUpdate(context.TODO(), r.client, prometheusRule, func(ignore runtime.Object) error {
		reconcilePrometheusRule(prometheusRule, instance)
		// Set UnifiedPushServer instance as the owner and controller
		err := controllerutil.SetControllerReference(instance, prometheusRule, r.scheme)
		return err
	})
	if err != nil {
		return r.manageError(instance, err)
	}
	if op != controllerutil.OperationResultNone {
		reqLogger.Info("PrometheusRule reconciled:", "PrometheusRule.Name", prometheusRule.Name, "PrometheusRule.Namespace", prometheusRule.Namespace, "Operation", op)
	}
	//## endregion PrometheusRule

	//## region GrafanaDasboard
	grafanaDashboard := &integreatlyv1alpha1.GrafanaDashboard{ObjectMeta: metav1.ObjectMeta{Name: "unifiedpushserver-dashboard", Namespace: instance.Namespace}}
	op, err = controllerutil.CreateOrUpdate(context.TODO(), r.client, grafanaDashboard, func(ignore runtime.Object) error {
		reconcileGrafanaDashboard(grafanaDashboard, instance)
		// Set UnifiedPushServer instance as the owner and controller
		err := controllerutil.SetControllerReference(instance, grafanaDashboard, r.scheme)
		return err
	})
	if err != nil {
		return r.manageError(instance, err)
	}
	if op != controllerutil.OperationResultNone {
		reqLogger.Info("GrafanaDashboard reconciled:", "GrafanaDashboard.Name", grafanaDashboard.Name, "GrafanaDashboard.Namespace", grafanaDashboard.Namespace, "Operation", op)
	}
	//## endregion GrafanaDasboard
	//#endregion

	return r.manageSuccess(instance, secondaryResources)
}

func (r *ReconcileUnifiedPushServer) manageError(instance *pushv1alpha1.UnifiedPushServer, issue error) (reconcile.Result, error) {
	r.recorder.Event(instance, "Warning", "ReconcileFailed", issue.Error())

	instance.Status.Message = issue.Error()
	instance.Status.Ready = false
	instance.Status.Phase = v1alpha1.PhaseFailing

	err := r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, "Unable to update status")
	}

	return reconcile.Result{
		RequeueAfter: requeueErrorDelay,
		Requeue:      true,
	}, nil
}

func (r *ReconcileUnifiedPushServer) manageSuccess(instance *pushv1alpha1.UnifiedPushServer, secondaryResources resources) (reconcile.Result, error) {
	instance.Status.Ready = r.readyStatus
	instance.Status.Message = ""
	instance.Status.SecondaryResources = secondaryResources

	// If resources are ready and we have not errored before now, we are in a reconciling phase
	if r.readyStatus {
		instance.Status.Phase = pushv1alpha1.PhaseReconciling
	} else {
		instance.Status.Phase = pushv1alpha1.PhaseInitializing
	}

	err := r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		log.Error(err, "Unable to update status")
		return reconcile.Result{
			RequeueAfter: requeueErrorDelay,
			Requeue:      true,
		}, nil
	}

	log.Info("Reconcile successful", "UnifiedPushServer.Namespace", instance.Namespace, "UnifiedPushServer.Name", instance.Name)
	return reconcile.Result{RequeueAfter: requeueDelay}, nil
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
