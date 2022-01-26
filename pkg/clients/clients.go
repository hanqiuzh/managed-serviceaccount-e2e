package clients

import (
	"k8s.io/client-go/dynamic"

	libgooptions "github.com/stolostron/library-e2e-go/pkg/options"
	libgoclient "github.com/stolostron/library-go/pkg/client"
)

func GetHubDynamicClient() (dynamic.Interface, error) {
	dynamicClient, err := libgoclient.NewKubeClientDynamic(
		libgooptions.TestOptions.Options.Hub.ApiServerURL,
		libgooptions.TestOptions.Options.Hub.KubeConfig,
		libgooptions.TestOptions.Options.Hub.KubeContext,
	)
	if err != nil {
		return nil, err
	}

	return dynamicClient, nil
}

func GetManagedClusterDynamicClient(managedClusterName string) (dynamic.Interface, error) {
	var targetManagedCluster libgooptions.Cluster
	optionManagedClusters := libgooptions.TestOptions.Options.ManagedClusters
	for _, cluster := range optionManagedClusters {
		if cluster.Name == managedClusterName {
			targetManagedCluster = cluster
			break
		}
	}

	dynamicClient, err := libgoclient.NewKubeClientDynamic(
		"",
		targetManagedCluster.KubeConfig,
		targetManagedCluster.KubeContext,
	)
	if err != nil {
		return nil, err
	}

	return dynamicClient, nil
}
