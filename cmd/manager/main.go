package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"github.com/aerogear/unifiedpush-operator/pkg/apis"
	"github.com/aerogear/unifiedpush-operator/pkg/controller"

	enmassev1beta "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1beta1"
	messaginguserv1beta "github.com/enmasseproject/enmasse/pkg/apis/user/v1beta1"
	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	openshiftappsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/aerogear/unifiedpush-operator/version"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monclientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Starting the UnifiedPush Operator with Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "unifiedpush-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.", "")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for OpenShift Route
	if err := routev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	//Watch AMQ Online Resources
	if err := messaginguserv1beta.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := enmassev1beta.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for OpenShift Apis
	if err := openshiftappsv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for OpenShift Image apis
	if err := imagev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for Monitoring apis
	if err := monitoringv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for Integreatly Grafana apis
	if err := integreatlyv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	operatorNamespace, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		log.Error(err, "")
	}

	if err = serveCRMetrics(cfg); err != nil {
		log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
	}

	// Add to the below struct any other metrics ports you want to expose.
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	}

	// Create Service object to expose the metrics port(s).
	service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		log.Info("Could not create metrics Service", "error", err.Error())
	}

	if service != nil {
		err = addMonitoringKeyLabelToService(cfg, operatorNamespace, service)
		if err != nil {
			log.Error(err, "Could not add monitoring-key label to operator metrics Service")
		}

		err = createServiceMonitor(cfg, operatorNamespace, service)
		if err != nil {
			log.Info("Could not create ServiceMonitor object", "error", err.Error())
			// If this operator is deployed to a cluster without the prometheus-operator running, it will return
			// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
			if err == metrics.ErrServiceMonitorNotPresent {
				log.Info("Install prometheus-operator in you cluster to create ServiceMonitor objects", "error", err.Error())
			}
		}
	}

	client := mgr.GetClient()
	prometheusRule := &monitoringv1.PrometheusRule{ObjectMeta: metav1.ObjectMeta{Name: "unifiedpush-operator", Namespace: operatorNamespace}}

	controllerutil.CreateOrUpdate(ctx, client, prometheusRule, func(ignore k8sruntime.Object) error {
		reconcilePrometheusRule(prometheusRule)
		return nil
	})

	grafanaDashboard := &integreatlyv1alpha1.GrafanaDashboard{ObjectMeta: metav1.ObjectMeta{Name: "unifiedpush-operator", Namespace: operatorNamespace}}

	controllerutil.CreateOrUpdate(ctx, client, grafanaDashboard, func(ignore k8sruntime.Object) error {
		reconcileGrafanaDashboard(grafanaDashboard)
		return nil
	})

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

func addMonitoringKeyLabelToService(cfg *rest.Config, ns string, service *v1.Service) error {
	kclient, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}

	updatedLabels := map[string]string{"monitoring-key": "middleware"}
	for k, v := range service.ObjectMeta.Labels {
		updatedLabels[k] = v
	}
	service.ObjectMeta.Labels = updatedLabels

	err = kclient.Update(context.TODO(), service)
	if err != nil {
		return err
	}

	return nil
}

// createServiceMonitor is a temporary fix until the version in the
// operator-sdk is fixed to have the correct Path set on the Endpoints
func createServiceMonitor(config *rest.Config, ns string, service *v1.Service) error {
	mclient := monclientv1.NewForConfigOrDie(config)

	sm := metrics.GenerateServiceMonitor(service)
	eps := []monitoringv1.Endpoint{}
	for _, ep := range sm.Spec.Endpoints {
		eps = append(eps, monitoringv1.Endpoint{Port: ep.Port, Path: "/metrics"})
	}
	sm.Spec.Endpoints = eps

	_, err := mclient.ServiceMonitors(ns).Create(sm)
	if err != nil {
		return err
	}

	return nil
}

// serveCRMetrics gets the Operator/CustomResource GVKs and generates metrics based on those types.
// It serves those metrics on "http://metricsHost:operatorMetricsPort".
func serveCRMetrics(cfg *rest.Config) error {
	// Below function returns filtered operator/CustomResource specific GVKs.
	// For more control override the below GVK list with your own custom logic.
	filteredGVK, err := k8sutil.GetGVKsFromAddToScheme(apis.AddToScheme)
	if err != nil {
		return err
	}
	// Get the namespace the operator is currently deployed in.
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}
	// To generate metrics in other namespaces, add the values below.
	ns := []string{operatorNs}
	// Generate and serve custom resource specific metrics.
	err = kubemetrics.GenerateAndServeCRMetrics(cfg, ns, filteredGVK, metricsHost, operatorMetricsPort)
	if err != nil {
		return err
	}
	return nil
}

func reconcilePrometheusRule(promethuesRule *monitoringv1.PrometheusRule) {
	labels := map[string]string{
		"monitoring-key": "middleware",
		"prometheus":     "application-monitoring",
		"role":           "alert-rules",
	}
	critical := map[string]string{
		"severity": "critical",
	}
	sop_url := fmt.Sprintf("https://github.com/aerogear/unifiedpush-operator/blob/%s/SOP/SOP-operator.adoc", version.Version)
	upsPushOperatorDown := map[string]string{
		"description": "The UnifiedPush Operator has been down for more than 5 minutes.",
		"summary":     "The UnifiedPush Operator is down.",
		"sop_url":     sop_url,
	}
	operatorName, err := k8sutil.GetOperatorName()
	if err != nil {
		log.Error(err, "")
	}
	promethuesRule.ObjectMeta.Labels = labels
	job := fmt.Sprintf("%s-metrics", operatorName)
	promethuesRule.Spec = monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{
			{
				Name: "general.rules",
				Rules: []monitoringv1.Rule{
					{
						Alert: "UnifiedPushOperatorDown",
						Expr: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: fmt.Sprintf("absent(up{service=\"%s\"} == 1)", job),
						},
						For:         "5m",
						Labels:      critical,
						Annotations: upsPushOperatorDown,
					},
				},
			},
		},
	}
}

func reconcileGrafanaDashboard(grafanaDashboard *integreatlyv1alpha1.GrafanaDashboard) {
	operatorNamespace, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		log.Error(err, "")
	}
	labels := map[string]string{
		"monitoring-key": "middleware",
		"prometheus":     "application-monitoring",
	}
	operatorName, err := k8sutil.GetOperatorName()
	if err != nil {
		log.Error(err, "")
	}
	service := fmt.Sprintf("%s-metrics", operatorName)

	grafanaDashboard.ObjectMeta.Labels = labels
	grafanaDashboard.Spec = integreatlyv1alpha1.GrafanaDashboardSpec{
		Name: "unifiedpushoperator.json",
		Json: `
		{
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
			"description": "Operator metrics",
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
				"fill": 1,
				"gridPos": {
				  "h": 8,
				  "w": 24,
				  "x": 3,
				  "y": 1
				},
				"id": 1,
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
					"expr": "kube_endpoint_address_available{namespace='` + operatorNamespace + `',endpoint='` + service + `'}",
					"format": "time_series",
					"hide": false,
					"intervalFactor": 2,
					"legendFormat": "{{ '{{' }}service{{ '}}' }} - Uptime",
					"metric": "",
					"refId": "A",
					"step": 2
				  }
				],
				"thresholds": [],
				"timeFrom": null,
				"timeRegions": [],
				"timeShift": null,
				"title": "Uptime",
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
					"format": "none",
					"label": null,
					"logBase": null,
					"max": 1.5,
					"min": 0,
					"show": true
				  },
				  {
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
				"collapsed": false,
				"gridPos": {
				  "h": 1,
				  "w": 24,
				  "x": 0,
				  "y": 9
				},
				"id": 10,
				"panels": [],
				"repeat": null,
				"title": "Resources",
				"type": "row"
			  },
			  {
				"aliasColors": {},
				"bars": false,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"fill": 1,
				"gridPos": {
				  "h": 8,
				  "w": 24,
				  "x": 0,
				  "y": 10
				},
				"id": 4,
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
					"expr": "process_virtual_memory_bytes{namespace='` + operatorNamespace + `',service='` + service + `'}",
					"format": "time_series",
					"intervalFactor": 1,
					"legendFormat": "Virtual Memory",
					"refId": "A"
				  },
				  {
					"expr": "process_resident_memory_bytes{namespace='` + operatorNamespace + `',service='` + service + `'}",
					"format": "time_series",
					"intervalFactor": 2,
					"legendFormat": "Memory Usage",
					"refId": "B",
					"step": 2
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
				"bars": false,
				"dashLength": 10,
				"dashes": false,
				"datasource": "Prometheus",
				"fill": 1,
				"gridPos": {
				  "h": 8,
				  "w": 24,
				  "x": 0,
				  "y": 18
				},
				"id": 2,
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
					"expr": "sum(rate(process_cpu_seconds_total{namespace='` + operatorNamespace + `',service='` + service + `'}[1m]))*1000",
					"format": "time_series",
					"interval": "",
					"intervalFactor": 2,
					"legendFormat": "UnifiedPush Operator- CPU Usage in Millicores",
					"refId": "A",
					"step": 2
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
					"format": "short",
					"label": "Millicores",
					"logBase": 10,
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
			  }
			],
			"refresh": "10s",
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
			"title": "UnifiedPush Operator",
			"version": 2
		  }
		`,
	}
}
