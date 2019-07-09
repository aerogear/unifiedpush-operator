package androidvariant

import (
	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	androidVariantInstance = pushv1alpha1.AndroidVariant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-androidvariant",
			Namespace: "unifiedpush",
		},
		Spec: pushv1alpha1.AndroidVariantSpec{
			Description:       "My super Android variant",
			ServerKey:         "somekeyinbase64==",
			SenderId:          "123456",
			PushApplicationId: "123456",
		},
	}
)
