package unifiedpushserver

import (
	"fmt"

	"github.com/aerogear/unifiedpush-operator/pkg/constants"
	"github.com/aerogear/unifiedpush-operator/version"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"

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

func buildEnv(cr *pushv1alpha1.UnifiedPushServer) []corev1.EnvVar {
	var env = []corev1.EnvVar{
		{
			Name: "POSTGRES_SERVICE_HOST",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "POSTGRES_HOST",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-postgresql", cr.Name),
					},
				},
			},
		},
		{
			Name: "POSTGRES_SERVICE_PORT",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "POSTGRES_PORT",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-postgresql", cr.Name),
					},
				},
			},
		},
		{
			Name: "POSTGRES_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "POSTGRES_USERNAME",
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
					Key: "POSTGRES_PASSWORD",
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
					Key: "POSTGRES_DATABASE",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-postgresql", cr.Name),
					},
				},
			},
		},
	}

	if cr.Spec.UseMessageBroker {
		env = append(env,
			corev1.EnvVar{
				Name:  "ARTEMIS_USER",
				Value: "upsuser",
			},

			corev1.EnvVar{
				Name: "ARTEMIS_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "artemis-password",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf("%s-amq", cr.Name),
						},
					},
				},
			},
			corev1.EnvVar{
				Name: "ARTEMIS_SERVICE_HOST",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						Key: "artemis-url",
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf("%s-amq", cr.Name),
						},
					},
				},
			},
			corev1.EnvVar{
				Name:  "ARTEMIS_SERVICE_PORT",
				Value: "5672",
			})
	}

	return env

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

	replicas := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cr.Name,
					InitContainers: []corev1.Container{
						{
							Name:            cfg.PostgresContainerName,
							Image:           constants.PostgresImage,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name: "POSTGRES_SERVICE_HOST",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "POSTGRES_HOST",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: fmt.Sprintf("%s-postgresql", cr.Name),
											},
										},
									},
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
							Name:            cfg.UPSContainerName,
							Image:           constants.UPSImage,
							ImagePullPolicy: corev1.PullAlways,
							Env:             buildEnv(cr),
							Resources:       getUnifiedPushResourceRequirements(cr),
							Ports: []corev1.ContainerPort{
								{
									Name:          cfg.UPSContainerName,
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
								InitialDelaySeconds: 120,
								TimeoutSeconds:      10,
							},
						},
						{
							Name:            cfg.OauthProxyContainerName,
							Image:           constants.OauthProxyImage,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "public",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 4180,
								},
							},
							Resources: getOauthProxyResourceRequirements(cr),
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

func reconcilePrometheusRule(prometheusRule *monitoringv1.PrometheusRule, cr *pushv1alpha1.UnifiedPushServer) {

	labels := map[string]string{
		"monitoring-key": "middleware",
	}
	critical := map[string]string{
		"severity": "critical",
	}
	warning := map[string]string{
		"severity": "warning",
	}
	sop_url := fmt.Sprintf("https://github.com/aerogear/unifiedpush-operator/blob/%s/SOP/SOP-push.adoc", version.Version)
	unifiedPushDownAnnotations := map[string]string{
		"description": "The aerogear-unifiedpush-server Pod Server has been down for more than 5 minutes.",
		"summary":     "The aerogear-unifiedpush-server Pod Server is down.",
		"sop_url":     sop_url,
	}
	unifiedPushConsoleDownAnnotations := map[string]string{
		"description": "The aerogear-unifiedpush-server admin console has been down for more than 5 minutes.",
		"summary":     "The aerogear-unifiedpush-server admin console endpoint has been unavailable for more that 5 minutes.",
		"sop_url":     sop_url,
	}
	unifiedPushDatabaseDownAnnotations := map[string]string{
		"description": "The aerogear-unifiedpush-server Database pod has been down for more than 5 minutes.",
		"summary":     "The aerogear-unifiedpush-server Database is down.",
		"sop_url":     sop_url,
	}
	unifiedPushJavaHeapThresholdExceededAnnotations := map[string]string{
		"description": "The Heap Usage of the aerogear-unifiedpush-server Server exceeded 90% of usage.",
		"summary":     "The aerogear-unifiedpush-server Server JVM Heap Threshold Exceeded 90% of usage.",
		"sop_url":     sop_url,
	}
	unifiedPushJavaNonHeapThresholdExceededAnnotations := map[string]string{
		"description": "The nonheap usage of the aerogear-unifiedpush-server Server exceeded 90% of usage.",
		"summary":     "The nonheap usage of the aerogear-unifiedpush-server Server exceeded 90% of usage.",
		"sop_url":     sop_url,
	}
	unifiedPushJavaGCTimePerMinuteScavengeAnnotations := map[string]string{
		"description": "Amount of time per minute spent on garbage collection in the aerogear-unifiedpush-server Server pod exceeds 90%",
		"summary":     "Amount of time per minute spent on garbage collection in the aerogear-unifiedpush-server Server pod exceeds 90%. This could indicate that the available heap memory is insufficient.",
		"sop_url":     sop_url,
	}
	unifiedPushJavaDeadlockedThreadsAnnotations := map[string]string{
		"description": "Number of threads in deadlock state of the aerogear-unifiedpush-server Server > 0.",
		"summary":     "Number of threads in deadlock state of the aerogear-unifiedpush-server Server > 0.",
		"sop_url":     sop_url,
	}
	unifiedPushMessagesFailuresAnnotations := map[string]string{
		"description": "More than 50 failed requests attempts for aerogear-unifiedpush-server Server fails over the last 5 minutes.",
		"summary":     "More than 50 failed messages attempts for aerogear-unifiedpush-server Server fails over the last 5 minutes.",
		"sop_url":     sop_url,
	}
	namespace := cr.Namespace
	upsEndpoint := fmt.Sprintf("%s-unifiedpush", cr.ObjectMeta.Name)
	prometheusRule.ObjectMeta.Labels = labels
	prometheusRule.Spec = monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{
			{
				Name: "general.rules",
				Rules: []monitoringv1.Rule{
					{
						Alert: "UnifiedPushDown",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: fmt.Sprintf("absent(kube_pod_container_status_running{namespace=\"%s\",container=\"ups\"} >= 1)", namespace),
						},
						For:         "5m",
						Labels:      critical,
						Annotations: unifiedPushDownAnnotations,
					},
					{
						Alert: "UnifiedPushConsoleDown",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: fmt.Sprintf("absent(kube_endpoint_address_available{endpoint=\"%s\"} == 1)", upsEndpoint),
						},
						For:         "5m",
						Labels:      critical,
						Annotations: unifiedPushConsoleDownAnnotations,
					},
					{
						Alert: "UnifiedPushJavaHeapThresholdExceeded",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: fmt.Sprintf("100 * jvm_memory_bytes_used{area=\"heap\",namespace=\"%s\"}/ jvm_memory_bytes_max{area=\"heap\",namespace=\"%s\"}> 90", namespace, namespace),
						},
						For:         "1m",
						Labels:      critical,
						Annotations: unifiedPushJavaHeapThresholdExceededAnnotations,
					},
					{
						Alert: "UnifiedPushJavaNonHeapThresholdExceeded",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: fmt.Sprintf("100 * jvm_memory_bytes_used{area=\"nonheap\",namespace=\"%s\"}/ jvm_memory_bytes_max{area=\"nonheap\",namespace=\"%s\"}> 90", namespace, namespace),
						},
						For:         "1m",
						Labels:      critical,
						Annotations: unifiedPushJavaNonHeapThresholdExceededAnnotations,
					},
					{
						Alert: "UnifiedPushJavaGCTimePerMinuteScavenge",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: fmt.Sprintf("increase(jvm_gc_collection_seconds_sum{namespace=\"%s\",service=\"%s\"}[1m]) > 1 * 60 * 0.9", namespace, upsEndpoint),
						},
						For:         "1m",
						Labels:      critical,
						Annotations: unifiedPushJavaGCTimePerMinuteScavengeAnnotations,
					},
					{
						Alert: "UnifiedPushJavaDeadlockedThreads",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: fmt.Sprintf("jvm_threads_deadlocked{namespace=\"%s\",service=\"%s\"}> 0", namespace, upsEndpoint),
						},
						For:         "1m",
						Labels:      warning,
						Annotations: unifiedPushJavaDeadlockedThreadsAnnotations,
					},
					{
						Alert: "UnifiedPushMessagesFailures",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: fmt.Sprintf("rate(aerogear_ups_push_requests_fail_total{namespace=\"%s\",service=\"%s\"}[5m])* 300 > 50", namespace, upsEndpoint),
						},
						For:         "5m",
						Labels:      warning,
						Annotations: unifiedPushMessagesFailuresAnnotations,
					},
				},
			},
		},
	}

	// Don't add UnifiedPushDatabaseDown rule if there's no Postgresql
	if !cr.Spec.ExternalDB {
		rule := monitoringv1.Rule{
			Alert: "UnifiedPushDatabaseDown",
			Expr: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: fmt.Sprintf("absent(kube_pod_container_status_running{namespace=\"%s\",container=\"postgresql\"} == 1)", namespace),
			},
			For:         "5m",
			Labels:      critical,
			Annotations: unifiedPushDatabaseDownAnnotations,
		}

		prometheusRule.Spec.Groups[0].Rules = append(prometheusRule.Spec.Groups[0].Rules, rule)
	}
}

func reconcileServiceMonitor(serviceMonitor *monitoringv1.ServiceMonitor) {
	labels := map[string]string{
		"monitoring-key": "middleware",
	}
	matchLabels := map[string]string{
		"internal": "unifiedpush",
	}
	serviceMonitor.ObjectMeta.Labels = labels
	serviceMonitor.Spec = monitoringv1.ServiceMonitorSpec{
		Endpoints: []monitoringv1.Endpoint{
			{
				Path: "/rest/prometheus/metrics",
				Port: "web",
			},
		},
		Selector: metav1.LabelSelector{
			MatchLabels: matchLabels,
		},
	}
}

func reconcileGrafanaDashboard(grafanaDashboard *integreatlyv1alpha1.GrafanaDashboard, cr *pushv1alpha1.UnifiedPushServer) {
	labels := map[string]string{
		"monitoring-key": "middleware",
		"prometheus":     "application-monitoring",
	}
	metaName := cr.ObjectMeta.Name
	upsEndpoint := fmt.Sprintf("%s-unifiedpush", metaName)
	postgresEndpoint := fmt.Sprintf("%s-postgresql", metaName)
	namespace := cr.Namespace
	grafanaDashboard.ObjectMeta.Labels = labels
	grafanaDashboard.Spec = integreatlyv1alpha1.GrafanaDashboardSpec{
		Name: "unifiedpushserver.json",
		Json: `{
			"__requires": [
				{
				"type": "grafana",
				"id": "grafana",
				"name": "Grafana",
				"version": "4.3.2"
				},
				{
				"type": "panel",
				"id": "graph",
				"name": "Graph",
				"version": ""
				},
				{
				"type": "datasource",
				"id": "prometheus",
				"name": "Prometheus",
				"version": "1.0.0"
				},
				{
				"type": "panel",
				"id": "singlestat",
				"name": "Singlestat",
				"version": ""
				}
			],
			"annotations": {
				"list": [
				{
				"builtIn": 1,
				"datasource": "-- Grafana --",
				"enable": true,
				"hide": true,
				"iconColor": "rgba(0, 211, 255, 1)",
				"name": "Annotations & Alerts",
				"type": "dashboard"
				}
				]
			},
			"description": "Application metrics",
			"editable": true,
			"gnetId": null,
			"graphTooltip": 0,
			"links": [],
			"panels": [
			{
				"collapsed": false,
				"gridPos": {
				"h": 1,
				"w": 24,
				"x": 0,
				"y": 0
				},
				"id": 9,
				"panels": [],
				"repeat": null,
				"title": "Uptime",
				"type": "row"
			},
			{
				"aliasColors": {},
				"bars": true,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"decimals": 0,
				"fill": 1,
				"gridPos": {
				"h": 8,
				"w": 24,
				"x": 3,
				"y": 1
				},
				"id": 1,
				"legend": {
				"alignAsTable": true,
				"avg": false,
				"current": false,
				"max": false,
				"min": false,
				"show": true,
				"total": false,
				"values": false
				},
				"lines": true,
				"linewidth": 1,
				"links": [
				{
				"type": "dashboard"
				}
				],
				"nullPointMode": "null",
				"percentage": true,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": false,
				"steppedLine": false,
				"targets": [
				{
				"expr": "kube_endpoint_address_available{namespace='` + namespace + `',endpoint='` + upsEndpoint + `'}",
				"format": "time_series",
				"hide": false,
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "uptime service endpoint",
				"metric": "",
				"refId": "A",
				"step": 2
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "Admin Console - Uptime",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"decimals": null,
				"format": "none",
				"label": null,
				"logBase": null,
				"max": 1.5,
				"min": 0,
				"show": true
				},
				{
				"decimals": null,
				"format": "short",
				"label": null,
				"logBase": null,
				"max": 2,
				"min": 0,
				"show": true
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			},
			{
				"aliasColors": {},
				"bars": true,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"fill": 1,
				"gridPos": {
				"h": 8,
				"w": 24,
				"x": 3,
				"y": 9
				},
				"id": 12,
				"legend": {
				"alignAsTable": true,
				"avg": false,
				"current": false,
				"max": false,
				"min": false,
				"show": true,
				"total": false,
				"values": false
				},
				"lines": true,
				"linewidth": 1,
				"links": [],
				"nullPointMode": "null",
				"percentage": false,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": false,
				"steppedLine": true,
				"targets": [
				{
				"expr": "kube_endpoint_address_available{namespace='` + namespace + `',endpoint='` + postgresEndpoint + `'}",
				"format": "time_series",
				"hide": false,
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "database container ",
				"refId": "A"
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "UnifiedPush Database - Uptime",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": true
				},
				{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": true
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			},
			{
				"aliasColors": {},
				"bars": true,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"fill": 1,
				"gridPos": {
				"h": 8,
				"w": 24,
				"x": 3,
				"y": 17
				},
				"id": 14,
				"legend": {
				"avg": false,
				"current": false,
				"max": false,
				"min": false,
				"show": true,
				"total": false,
				"values": false
				},
				"lines": false,
				"linewidth": 1,
				"links": [],
				"nullPointMode": "null",
				"percentage": false,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": false,
				"steppedLine": true,
				"targets": [
				{
				"expr": "kube_endpoint_address_available{namespace='` + namespace + `',endpoint='` + upsEndpoint + `'}",
				"format": "time_series",
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "server container",
				"refId": "A"
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "UnifiedPush Server  - Uptime",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": true
				},
				{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": true
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			},
			{
				"collapsed": false,
				"gridPos": {
				"h": 1,
				"w": 24,
				"x": 0,
				"y": 25
				},
				"id": 10,
				"panels": [],
				"repeat": null,
				"title": "Resources",
				"type": "row"
			},
			{
				"aliasColors": {},
				"bars": true,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"fill": 1,
				"gridPos": {
				"h": 8,
				"w": 24,
				"x": 0,
				"y": 26
				},
				"id": 40,
				"interval": "1s",
				"legend": {
				"alignAsTable": true,
				"avg": true,
				"current": false,
				"max": true,
				"min": true,
				"show": true,
				"total": false,
				"values": true
				},
				"lines": true,
				"linewidth": 1,
				"links": [],
				"nullPointMode": "null",
				"percentage": false,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": false,
				"steppedLine": true,
				"targets": [
				{
				"expr": "max(jvm_memory_bytes_used{area='heap',namespace='` + namespace + `'}) + max(jvm_memory_bytes_used{area='nonheap', namespace='` + namespace + `'})",
				"format": "time_series",
				"hide": false,
				"instant": false,
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "JVM memory",
				"refId": "A"
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "Current Memory",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"format": "bytes",
				"label": null,
				"logBase": 32,
				"max": null,
				"min": null,
				"show": true
				},
				{
				"format": "short",
				"label": null,
				"logBase": 32,
				"max": null,
				"min": null,
				"show": true
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			},
			{
				"aliasColors": {},
				"bars": true,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"fill": 1,
				"gridPos": {
				"h": 8,
				"w": 24,
				"x": 0,
				"y": 34
				},
				"id": 41,
				"interval": "1s",
				"legend": {
				"alignAsTable": true,
				"avg": true,
				"current": false,
				"max": true,
				"min": true,
				"show": true,
				"total": false,
				"values": true
				},
				"lines": true,
				"linewidth": 1,
				"links": [],
				"nullPointMode": "null",
				"percentage": false,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": false,
				"steppedLine": true,
				"targets": [
				{
				"expr": "sum(jvm_memory_bytes_max{namespace='` + namespace + `'})",
				"format": "time_series",
				"hide": false,
				"instant": false,
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "max",
				"refId": "A"
				},
				{
				"expr": "sum(jvm_memory_bytes_committed{namespace='` + namespace + `'})",
				"format": "time_series",
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "commited",
				"refId": "B"
				},
				{
				"expr": "sum(jvm_memory_bytes_used{namespace='` + namespace + `'})",
				"format": "time_series",
				"hide": false,
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "used",
				"refId": "C"
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "Memory Usage",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"format": "bytes",
				"label": null,
				"logBase": 32,
				"max": null,
				"min": null,
				"show": true
				},
				{
				"format": "short",
				"label": null,
				"logBase": 32,
				"max": null,
				"min": null,
				"show": true
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			},
			{
				"aliasColors": {},
				"bars": true,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"fill": 1,
				"gridPos": {
				"h": 8,
				"w": 24,
				"x": 0,
				"y": 42
				},
				"id": 42,
				"interval": "1s",
				"legend": {
				"alignAsTable": true,
				"avg": true,
				"current": false,
				"max": true,
				"min": true,
				"show": true,
				"total": false,
				"values": true
				},
				"lines": true,
				"linewidth": 1,
				"links": [],
				"nullPointMode": "null",
				"percentage": false,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": false,
				"steppedLine": true,
				"targets": [
				{
				"expr": "sum(process_cpu_seconds_total{namespace='` + namespace + `'})",
				"format": "time_series",
				"hide": false,
				"instant": false,
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "used",
				"refId": "A"
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "CPU Usage",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"format": "mbytes",
				"label": null,
				"logBase": 32,
				"max": null,
				"min": null,
				"show": true
				},
				{
				"format": "short",
				"label": null,
				"logBase": 32,
				"max": null,
				"min": null,
				"show": true
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			},
			{
				"collapsed": false,
				"gridPos": {
				"h": 1,
				"w": 24,
				"x": 0,
				"y": 50
				},
				"id": 20,
				"panels": [],
				"title": "API Monitoring",
				"type": "row"
			},
			{
				"aliasColors": {},
				"bars": false,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"decimals": 0,
				"fill": 1,
				"gridPos": {
				"h": 8,
				"w": 12,
				"x": 0,
				"y": 51
				},
				"id": 35,
				"interval": "1m",
				"legend": {
				"alignAsTable": false,
				"avg": true,
				"current": false,
				"hideEmpty": false,
				"max": false,
				"min": true,
				"rightSide": false,
				"show": true,
				"total": false,
				"values": true
				},
				"lines": true,
				"linewidth": 1,
				"links": [],
				"nullPointMode": "null",
				"percentage": false,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": true,
				"steppedLine": false,
				"targets": [
				{
				"expr": "sum(aerogear_ups_push_requests_total{namespace='` + namespace + `',service='` + upsEndpoint + `'})",
				"format": "time_series",
				"hide": false,
				"instant": false,
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "attempts",
				"refId": "A"
				},
				{
				"expr": "sum(aerogear_ups_push_requests_fail_total{namespace='` + namespace + `',service='` + upsEndpoint + `'})",
				"format": "time_series",
				"hide": false,
				"instant": false,
				"interval": "1s",
				"intervalFactor": 1,
				"legendFormat": "failures",
				"refId": "B"
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "Push Notifications",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"decimals": 0,
				"format": "short",
				"label": null,
				"logBase": 10,
				"max": null,
				"min": 0,
				"show": true
				},
				{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": false
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			},
			{
				"aliasColors": {},
				"bars": false,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"decimals": 0,
				"fill": 1,
				"gridPos": {
				"h": 8,
				"w": 12,
				"x": 12,
				"y": 51
				},
				"id": 39,
				"interval": "1m",
				"legend": {
				"alignAsTable": false,
				"avg": false,
				"current": false,
				"hideEmpty": false,
				"max": false,
				"min": false,
				"rightSide": false,
				"show": true,
				"total": false,
				"values": false
				},
				"lines": true,
				"linewidth": 1,
				"links": [],
				"nullPointMode": "null",
				"percentage": false,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": true,
				"steppedLine": false,
				"targets": [
				{
				"expr": "sum(aerogear_ups_push_requests_ios{namespace='` + namespace + `',service='` + upsEndpoint + `'})",
				"format": "time_series",
				"hide": false,
				"interval": "1m",
				"intervalFactor": 1,
				"legendFormat": "ios",
				"refId": "C"
				},
				{
				"expr": "sum(aerogear_ups_push_requests_android{namespace='` + namespace + `',service='` + upsEndpoint + `'})",
				"format": "time_series",
				"hide": false,
				"interval": "1m",
				"intervalFactor": 1,
				"legendFormat": "android",
				"refId": "D"
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "Push Notifications per Platform",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"decimals": 0,
				"format": "short",
				"label": null,
				"logBase": 10,
				"max": null,
				"min": 0,
				"show": true
				},
				{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": false
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			},
			{
				"aliasColors": {},
				"bars": false,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"fill": 1,
				"gridPos": {
				"h": 9,
				"w": 24,
				"x": 0,
				"y": 59
				},
				"id": 36,
				"interval": "1m",
				"legend": {
				"avg": false,
				"current": false,
				"max": false,
				"min": false,
				"show": true,
				"total": false,
				"values": false
				},
				"lines": true,
				"linewidth": 1,
				"links": [],
				"nullPointMode": "null",
				"percentage": false,
				"pointradius": 5,
				"points": false,
				"renderer": "flot",
				"seriesOverrides": [],
				"spaceLength": 10,
				"stack": false,
				"steppedLine": false,
				"targets": [
				{
				"expr": "aerogear_ups_device_register_requests_total{namespace='` + namespace + `',service='` + upsEndpoint + `'}",
				"format": "time_series",
				"hide": false,
				"interval": "",
				"intervalFactor": 1,
				"legendFormat": "requests to register devices",
				"refId": "A"
				}
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "Devices",
				"tooltip": {
				"shared": true,
				"sort": 0,
				"value_type": "individual"
				},
				"transparent": false,
				"type": "graph",
				"xaxis": {
				"buckets": null,
				"mode": "time",
				"name": null,
				"show": true,
				"values": []
				},
				"yaxes": [
				{
				"format": "bytes",
				"label": null,
				"logBase": 2,
				"max": null,
				"min": 0,
				"show": true
				},
				{
				"format": "short",
				"label": null,
				"logBase": 1,
				"max": null,
				"min": null,
				"show": false
				}
				],
				"yaxis": {
				"align": false,
				"alignLevel": null
				}
			}
			],
			"refresh": false,
			"schemaVersion": 16,
			"style": "dark",
			"tags": [],
			"templating": {
				"list": []
			},
			"time": {
				"from": "now/d",
				"to": "now"
			},
			"timepicker": {
				"refresh_intervals": [
				"5s",
				"10s",
				"30s",
				"1m",
				"5m",
				"15m",
				"30m",
				"1h",
				"2h",
				"1d"
				],
				"time_options": [
				"5m",
				"15m",
				"1h",
				"6h",
				"12h",
				"24h",
				"2d",
				"7d",
				"30d"
				]
			},
			"timezone": "browser",
			"title": "UnifiedPush Server",
			"version": 1
			}`,
	}
}
