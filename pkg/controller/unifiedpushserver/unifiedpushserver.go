package unifiedpushserver

import (
	"fmt"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/pkg/errors"
)

func newUnifiedPushServiceAccount(cr *pushv1alpha1.UnifiedPushServer) (*corev1.ServiceAccount, error) {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				"serviceaccounts.openshift.io/oauth-redirectreference.ups": fmt.Sprintf("{\"kind\":\"OAuthRedirectReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"Route\",\"name\":\"%s-unifiedpush-proxy\"}}", cr.Name),
			},
		},
	}, nil
}

func newOauthProxyService(cr *pushv1alpha1.UnifiedPushServer) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: objectMeta(cr, "unifiedpush-proxy"),
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     cr.Name,
				"service": "ups",
			},
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4180,
					},
				},
			},
		},
	}, nil
}

func newOauthProxyRoute(cr *pushv1alpha1.UnifiedPushServer) (*routev1.Route, error) {
	return &routev1.Route{
		ObjectMeta: objectMeta(cr, "unifiedpush-proxy"),
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: fmt.Sprintf("%s-%s", cr.Name, "unifiedpush-proxy"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyNone,
			},
		},
	}, nil
}

func newUnifiedPushServerDeployment(cr *pushv1alpha1.UnifiedPushServer) (*appsv1.Deployment, error) {
	labels := map[string]string{
		"app":     cr.Name,
		"service": "ups",
	}

	cookieSecret, err := generatePassword()
	if err != nil {
		return nil, errors.Wrap(err, "error generating cookie secret")
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cr.Name,
					InitContainers: []corev1.Container{
						{
							Name:            POSTGRES_CONTAINER_NAME,
							Image:           postgresql.image(),
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRES_SERVICE_HOST",
									Value: fmt.Sprintf("%s-postgresql", cr.Name),
								},
							},
							Command: []string{
								"/bin/sh",
								"-c",
								"source /opt/rh/rh-postgresql96/enable && until pg_isready -h $POSTGRES_SERVICE_HOST; do echo waiting for database; sleep 2; done;",
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            UPS_CONTAINER_NAME,
							Image:           unifiedpush.image(),
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRES_SERVICE_HOST",
									Value: fmt.Sprintf("%s-postgresql", cr.Name),
								},
								{
									Name:  "POSTGRES_SERVICE_PORT",
									Value: "5432",
								},
								{
									Name: "POSTGRES_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "database-user",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: fmt.Sprintf("%s-postgresql", cr.Name),
											},
										},
									},
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "database-password",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: fmt.Sprintf("%s-postgresql", cr.Name),
											},
										},
									},
								},
								{
									Name: "POSTGRES_DATABASE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "database-name",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: fmt.Sprintf("%s-postgresql", cr.Name),
											},
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          UPS_CONTAINER_NAME,
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/rest/applications",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 8080,
										},
									},
								},
								InitialDelaySeconds: 15,
								TimeoutSeconds:      2,
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/rest/applications",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 8080,
										},
									},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      2,
							},
						},
						{
							Name:            OAUTH_PROXY_CONTAINER_NAME,
							Image:           proxy.image(),
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "public",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 4180,
								},
							},
							Args: []string{
								"--provider=openshift",
								fmt.Sprintf("--openshift-service-account=%s", cr.Name),
								"--upstream=http://localhost:8080",
								"--http-address=0.0.0.0:4180",
								"--skip-auth-regex=/rest/sender,/rest/registry/device,/rest/prometheus/metrics,/rest/auth/config",
								"--https-address=",
								fmt.Sprintf("--cookie-secret=%s", cookieSecret),
							},
						},
					},
				},
			},
		},
	}, nil
}

func newUnifiedPushServerService(cr *pushv1alpha1.UnifiedPushServer) (*corev1.Service, error) {
	serviceObjectMeta := objectMeta(cr, "unifiedpush")
	serviceObjectMeta.Annotations = map[string]string{
		"org.aerogear.metrics/plain_endpoint": "/rest/prometheus/metrics",
	}
	serviceObjectMeta.Labels["mobile"] = "enabled"
	serviceObjectMeta.Labels["internal"] = "unifiedpush"

	return &corev1.Service{
		ObjectMeta: serviceObjectMeta,
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     cr.Name,
				"service": "ups",
			},
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
		},
	}, nil
}
