package e2e

import (
	"testing"
	"time"

	apis "github.com/aerogear/unifiedpush-operator/pkg/apis"
	pushv1alpha1 "github.com/aerogear/unifiedpush-operator/pkg/apis/push/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 200
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestUnifiedpush(t *testing.T) {
	unifiedpushList := &pushv1alpha1.UnifiedPushServerList{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, unifiedpushList); err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	t.Run("unifiedpush-e2e", UnifiedpushTest)
}

func UnifiedpushTest(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: cleanupRetryInterval,
	})

	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Successfully initialized cluster resources")

	namespace, err := ctx.GetNamespace()

	f := framework.Global
	if err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "unifiedpush-operator", 1, retryInterval, timeout); err != nil {
		t.Fatal(err)
	}

}
