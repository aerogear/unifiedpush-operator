package unifiedpushserver

// This is mostly copied from
// https://github.com/keycloak/keycloak-operator/blob/7.0.1/pkg/common/readiness_checks.go

import (
	"errors"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

const (
	conditionStatusSuccess = "True"
)

func isRouteReady(route *routev1.Route) bool {
	if route == nil {
		return false
	}
	// A route has a an array of Ingress where each have an array of conditions
	for _, ingress := range route.Status.Ingress {
		for _, condition := range ingress.Conditions {
			// A successful route will have the admitted condition type as true
			if condition.Type == routev1.RouteAdmitted && condition.Status != conditionStatusSuccess {
				return false
			}
		}
	}
	return true
}

func isDeploymentReady(deployment *appsv1.Deployment) (bool, error) {
	if deployment == nil {
		return false, nil
	}
	// A deployment has an array of conditions
	for _, condition := range deployment.Status.Conditions {
		// One failure condition exists, if this exists, return the Reason
		if condition.Type == appsv1.DeploymentReplicaFailure {
			return false, errors.New(condition.Reason)
			// A successful deployment will have the progressing condition type as true
		} else if condition.Type == appsv1.DeploymentProgressing && condition.Status != conditionStatusSuccess {
			return false, nil
		}
	}
	return true, nil
}

func isJobReady(job *batchv1.Job) (bool, error) {
	if job == nil {
		return false, nil
	}

	return job.Status.Succeeded == 1, nil
}
