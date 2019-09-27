package webpushvariant

import (
	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	webpushVariantInstance = pushv1alpha1.WebPushVariant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-webpushvariant",
			Namespace: "unifiedpush",
		},
		Spec: pushv1alpha1.WebPushVariantSpec{
			Description:       "My super WebPush variant",
			PublicKey:         "BIk8YK3iWC3BfMt3GLEghzY4v5GwaZsTWKxDKm-FZry3Nx2E_q-4VW3501DkQ5TX1Pe7c3yIsajUk9hQAo3sT-0",
			PrivateKey:        "FTg6q0-BXP6m-i6cNpg8P6JKccCUwWaD4yuirotxqXo",
			PushApplicationId: "123456",
		},
	}
)
