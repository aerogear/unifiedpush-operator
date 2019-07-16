package unifiedpushserver

import (
	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	pushServerInstance = pushv1alpha1.UnifiedPushServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-unifiedpushserver",
			Namespace: "unifiedpush",
		},
	}
)
