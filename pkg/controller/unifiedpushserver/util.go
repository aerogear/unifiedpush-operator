package unifiedpushserver

import (
	"fmt"
	"strings"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	"github.com/google/uuid"
	openshiftappsv1 "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func labels(cr *pushv1alpha1.UnifiedPushServer, suffix string) map[string]string {
	return map[string]string{
		"app":     cr.Name,
		"service": fmt.Sprintf("%s-%s", cr.Name, suffix),
	}
}

// objectMeta returns the default ObjectMeta for all the other objects here
func objectMeta(cr *pushv1alpha1.UnifiedPushServer, suffix string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", cr.Name, suffix),
		Namespace: cr.Namespace,
		Labels:    labels(cr, suffix),
	}
}

func generatePassword() (string, error) {
	generatedPassword, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "error generating password")
	}
	return strings.Replace(generatedPassword.String(), "-", "", -1), nil
}

func findContainerSpec(deployment *openshiftappsv1.DeploymentConfig, name string) *corev1.Container {
	if deployment == nil || &deployment.Spec == nil || &deployment.Spec.Template == nil || &deployment.Spec.Template.Spec == nil || &deployment.Spec.Template.Spec.Containers == nil || len(deployment.Spec.Template.Spec.Containers) == 0 {
		return nil
	}

	for _, spec := range deployment.Spec.Template.Spec.Containers {
		if spec.Name == name {
			return &spec
		}
	}

	return nil
}

func updateContainerSpecImage(deployment *openshiftappsv1.DeploymentConfig, name string, image string) {
	if deployment == nil || &deployment.Spec == nil || &deployment.Spec.Template == nil || &deployment.Spec.Template.Spec == nil || &deployment.Spec.Template.Spec.Containers == nil || len(deployment.Spec.Template.Spec.Containers) == 0 {
		return
	}

	for idx, spec := range deployment.Spec.Template.Spec.Containers {
		if spec.Name == name {
			deployment.Spec.Template.Spec.Containers[idx].Image = image
		}
	}
}
