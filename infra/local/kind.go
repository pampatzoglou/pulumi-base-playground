package local

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// KindCluster sets up a KIND cluster using Pulumi
func KindCluster(ctx *pulumi.Context, clusterConfigPath string, clusterName string) (string, error) {
	// Step 1: Create the KIND cluster using a shell command
	createCluster, err := local.NewCommand(ctx, "createKindCluster", &local.CommandArgs{
		Create: pulumi.String(fmt.Sprintf("kind create cluster --config %s --name %s", clusterConfigPath, clusterName)),
		Delete: pulumi.String(fmt.Sprintf("kind delete cluster --name %s", clusterName)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create KIND cluster: %w", err)
	}

	// Step 2: Export kubeconfig path (assumes default KIND kubeconfig setup)
	kubeconfigPath := fmt.Sprintf("$HOME/.kube/config")
	ctx.Export("kubeconfig", pulumi.String(kubeconfigPath))

	return kubeconfigPath, nil
}
