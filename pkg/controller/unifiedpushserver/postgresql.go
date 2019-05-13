package unifiedpushserver

import (
	"fmt"

	aerogearv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/aerogear/v1alpha1"

	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newPostgresqlPersistentVolumeClaim(cr *aerogearv1alpha1.UnifiedPushServer) (*corev1.PersistentVolumeClaim, error) {
	pvcSize, err := resource.ParseQuantity("1Gi")
	if err != nil {
		return nil, errors.Wrap(err, "error parsing PostgreSQL PVC storage size")
	}

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: objectMeta(cr, "postgresql"),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: pvcSize,
				},
			},
		},
	}, nil
}

func newPostgresqlSecret(cr *aerogearv1alpha1.UnifiedPushServer) (*corev1.Secret, error) {
	databasePassword, err := generatePassword()
	if err != nil {
		return nil, err
	}

	return &corev1.Secret{
		ObjectMeta: objectMeta(cr, "postgresql"),
		StringData: map[string]string{
			"database-name":     "unifiedpush",
			"database-user":     "unifiedpush",
			"database-password": databasePassword,
		},
	}, nil
}

func newPostgresqlDeployment(cr *aerogearv1alpha1.UnifiedPushServer) (*appsv1.Deployment, error) {
	memoryLimit, err := resource.ParseQuantity("512Mi")
	if err != nil {
		return nil, errors.Wrap(err, "error parsing PostgreSQL container memory limit size")
	}

	return &appsv1.Deployment{
		ObjectMeta: objectMeta(cr, "postgresql"),
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels(cr, "postgresql"),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels(cr, "postgresql"),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "postgresql",
							Image:           postgresql.image(),
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name: "POSTGRESQL_USER",
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
									Name: "POSTGRESQL_PASSWORD",
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
									Name: "POSTGRESQL_DATABASE",
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
									Name:          "postgresql",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 5432,
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/sh",
											"-i",
											"-c",
											"psql -h 127.0.0.1 -U $POSTGRESQL_USER -q -d $POSTGRESQL_DATABASE -c 'SELECT 1'",
										},
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      1,
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 5432,
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: memoryLimit,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      fmt.Sprintf("%s-postgresql-data", cr.Name),
									MountPath: "/var/lib/pgsql/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: fmt.Sprintf("%s-postgresql-data", cr.Name),
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: fmt.Sprintf("%s-postgresql", cr.Name),
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func newPostgresqlService(cr *aerogearv1alpha1.UnifiedPushServer) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: objectMeta(cr, "postgresql"),
		Spec: corev1.ServiceSpec{
			Selector: labels(cr, "postgresql"),
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:     "postgresql",
					Protocol: corev1.ProtocolTCP,
					Port:     5432,
				},
			},
		},
	}, nil
}
