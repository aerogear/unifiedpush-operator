package util

import (
	"context"

	"github.com/go-logr/logr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AddFinalizer will add a finalizer to the PushApplication CR so that
// we can delete from UPS appropriately
func AddFinalizer(client client.Client, reqLogger logr.Logger, o metav1.Object) error {
	if len(o.GetFinalizers()) < 1 && o.GetDeletionTimestamp() == nil {
		reqLogger.Info("Adding Finalizer to the PushApplication")
		o.SetFinalizers([]string{"finalizer.push.aerogear.org"})

		// Update CR
		switch typedObject := o.(type) {
		case runtime.Object:
			err := client.Update(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Error(err, "Failed to update a CR with a finalizer")
				return err
			}
		default:
			reqLogger.Info("Can't determine the type of thing to add finalizer to")
		}
	}
	return nil
}
