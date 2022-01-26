package utils

import (
	libgooptions "github.com/stolostron/library-e2e-go/pkg/options"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

func toUnstructured(
	obj interface{},
) (*unstructured.Unstructured, error) {
	u, err := runtime.
		DefaultUnstructuredConverter.
		ToUnstructured(obj)

	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: u}, nil
}

func GetImportedCluster(hubClient dynamic.Interface) (*clusterv1.ManagedCluster, error) {
	for _, optionsManagedCluster := range libgooptions.TestOptions.Options.ManagedClusters {
		clusterName := optionsManagedCluster.Name
		// make sure that the cluster is already imported
		managedCluster, err := GetManagedCluster(hubClient, clusterName)
		if err != nil {
			continue
		}

		return managedCluster, nil
	}

	//TODO: throw error?
	return nil, nil
}
