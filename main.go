package main

import (
	"pulumi-project/infra/local"
	"pulumi-project/infra/aws"
	"pulumi-project/provider"
	"pulumi-project/k8s-resources"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration
		clusterType := ctx.Config("cluster:type")
		clusterName := ctx.Config("cluster:name")
		kubeconfig := ""

		// Handle local KIND cluster setup
		if clusterType == "local" {
			var err error
			kubeconfig, err = local.KindCluster(ctx, "./infra/local/kind-cluster.yaml", clusterName)
			if err != nil {
				return err
			}
		} else if clusterType == "aws" {
			// Handle EKS setup
			cluster, err := aws.EksCluster(ctx)
			if err != nil {
				return err
			}
			kubeconfig = cluster.Kubeconfig
		} else {
			return pulumi.Errorf("unsupported cluster type: %s", clusterType)
		}

		// Set up Kubernetes provider
		k8sProvider, err := provider.KubernetesProvider(ctx, kubeconfig)
		if err != nil {
			return err
		}

		// Deploy Kubernetes resources
		_, err = k8sresources.CreateNamespace(ctx, "my-namespace", k8sProvider)
		if err != nil {
			return err
		}

		err = k8sresources.DeployHelmChart(ctx, "my-chart", "./charts/my-chart-1.tgz", "./charts/values/my-chart-1-values.yaml", k8sProvider)
		if err != nil {
			return err
		}

		return nil
	})
}
