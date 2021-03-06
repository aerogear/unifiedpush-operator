// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServer":       schema_pkg_apis_push_v1alpha1_UnifiedPushServer(ref),
		"github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerSpec":   schema_pkg_apis_push_v1alpha1_UnifiedPushServerSpec(ref),
		"github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerStatus": schema_pkg_apis_push_v1alpha1_UnifiedPushServerStatus(ref),
	}
}

func schema_pkg_apis_push_v1alpha1_UnifiedPushServer(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "UnifiedPushServer is the Schema for the unifiedpushservers API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerSpec", "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_push_v1alpha1_UnifiedPushServerSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "UnifiedPushServerSpec defines the desired state of UnifiedPushServer",
				Properties: map[string]spec.Schema{
					"externalDB": {
						SchemaProps: spec.SchemaProps{
							Description: "ExternalDB can be set to true to use details from Database and connect to external db",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"database": {
						SchemaProps: spec.SchemaProps{
							Description: "Database allows specifying the external PostgreSQL details directly in the CR. Only one of Database or DatabaseSecret should be specified, and ExternalDB must be true, otherwise a new PostgreSQL instance will be created (and deleted) on the cluster automatically.",
							Ref:         ref("github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerDatabase"),
						},
					},
					"databaseSecret": {
						SchemaProps: spec.SchemaProps{
							Description: "DatabaseSecret allows reading the external PostgreSQL details from a pre-existing Secret (ExternalDB must be true for it to be used). Only one of Database or DatabaseSecret should be specified, and ExternalDB must be true, otherwise a new PostgreSQL instance will be created (and deleted) on the cluster automatically.\n\nHere's an example of all of the fields that the secret must contain:\n\nPOSTGRES_DATABASE: sampledb POSTGRES_HOST: 172.30.139.148 POSTGRES_PORT: \"5432\" POSTGRES_USERNAME: userMSM POSTGRES_PASSWORD: RmwWKKIM7or7oJig POSTGRES_SUPERUSER: \"false\" POSTGRES_VERSION: \"10\"",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"backups": {
						SchemaProps: spec.SchemaProps{
							Description: "Backups is an array of configs that will be used to create CronJob resource instances",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerBackup"),
									},
								},
							},
						},
					},
					"useMessageBroker": {
						SchemaProps: spec.SchemaProps{
							Description: "UseMessageBroker can be set to true to use managed queues, if you are using enmasse. Defaults to false.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"unifiedPushResourceRequirements": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.ResourceRequirements"),
						},
					},
					"oAuthResourceRequirements": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.ResourceRequirements"),
						},
					},
					"postgresResourceRequirements": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.ResourceRequirements"),
						},
					},
					"postgresPVCSize": {
						SchemaProps: spec.SchemaProps{
							Description: "PVC size for Postgres service",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"affinity": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.Affinity"),
						},
					},
					"tolerations": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.Toleration"),
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerBackup", "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1.UnifiedPushServerDatabase", "k8s.io/api/core/v1.Affinity", "k8s.io/api/core/v1.ResourceRequirements", "k8s.io/api/core/v1.Toleration"},
	}
}

func schema_pkg_apis_push_v1alpha1_UnifiedPushServerStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "UnifiedPushServerStatus defines the observed state of UnifiedPushServer",
				Properties: map[string]spec.Schema{
					"phase": {
						SchemaProps: spec.SchemaProps{
							Description: "Phase indicates whether the CR is reconciling(good), failing(bad), or initializing.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"message": {
						SchemaProps: spec.SchemaProps{
							Description: "Message is a more human-readable message indicating details about current phase or error.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"ready": {
						SchemaProps: spec.SchemaProps{
							Description: "Ready is True if all resources are in a ready state and all work is done (phase should be \"reconciling\"). The type in the Go code here is deliberately a pointer so that we can distinguish between false and \"not set\", since it's an optional field.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"secondaryResources": {
						SchemaProps: spec.SchemaProps{
							Description: "SecondaryResources is a map of all the secondary resources types and names created for this CR.  e.g \"Deployment\": [ \"DeploymentName1\", \"DeploymentName2\" ]",
							Type:        []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: []string{"array"},
										Items: &spec.SchemaOrArray{
											Schema: &spec.Schema{
												SchemaProps: spec.SchemaProps{
													Type:   []string{"string"},
													Format: "",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Required: []string{"phase"},
			},
		},
		Dependencies: []string{},
	}
}
