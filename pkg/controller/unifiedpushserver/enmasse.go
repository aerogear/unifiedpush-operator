package unifiedpushserver

import (
	"fmt"
	"strings"

	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	enmassev1beta "github.com/enmasseproject/enmasse/pkg/apis/enmasse/v1beta1"
	messaginguserv1beta "github.com/enmasseproject/enmasse/pkg/apis/user/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newQueue(cr *pushv1alpha1.UnifiedPushServer, address string) *enmassev1beta.Address {
	name := fmt.Sprintf("ups.%s", strings.ToLower(address))
	return &enmassev1beta.Address{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name,
			},
		},
		Spec: enmassev1beta.AddressSpec{
			Address: address,
			Type:    "queue",
			Plan:    "brokered-queue",
		},
	}
}

func newTopic(cr *pushv1alpha1.UnifiedPushServer, address string) *enmassev1beta.Address {

	name := fmt.Sprintf("ups.%s", strings.ToLower(strings.Replace(address, "topic/", "", 1))) //a topic has a prefix.
	return &enmassev1beta.Address{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": cr.Name,
			},
		},
		Spec: enmassev1beta.AddressSpec{
			Address: address,
			Type:    "topic",
			Plan:    "brokered-topic",
		},
	}
}

func newMessagingUser(cr *pushv1alpha1.UnifiedPushServer) *messaginguserv1beta.MessagingUser {
	password := []byte("password")

	return &messaginguserv1beta.MessagingUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ups.upsuser",
			Namespace: cr.Namespace,
			Labels:    labels(cr, "ups.upsuser"),
		},
		Spec: messaginguserv1beta.MessagingUserSpec{
			Username: "upsuser",
			Authentication: messaginguserv1beta.AuthenticationSpec{
				Type:     "password",
				Password: password,
			},
			Authorization: []messaginguserv1beta.AuthorizationSpec{
				messaginguserv1beta.AuthorizationSpec{
					Addresses: []string{
						"*",
					},
					Operations: []string{
						"send",
						"recv",
					},
				},
			},
		},
	}
}

func newAddressSpace(cr *pushv1alpha1.UnifiedPushServer) *enmassev1beta.AddressSpace {
	return &enmassev1beta.AddressSpace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ups",
			Namespace: cr.Namespace,
			Labels:    labels(cr, "ups"),
		},
		Spec: enmassev1beta.AddressSpaceSpec{
			Type: "brokered",
			Plan: "brokered-single-broker",
		},
	}
}