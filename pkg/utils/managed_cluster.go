package utils

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

func unstructuredToManagedCluster(
	u *unstructured.Unstructured,
) (*clusterv1.ManagedCluster, error) {
	mc := &clusterv1.ManagedCluster{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		u.UnstructuredContent(),
		mc,
	)
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func GetManagedCluster(
	hubClient dynamic.Interface,
	clusterName string,
) (*clusterv1.ManagedCluster, error) {
	gvr := schema.GroupVersionResource{
		Group:    "cluster.open-cluster-management.io",
		Version:  "v1",
		Resource: "managedclusters",
	}

	uManagedCluster, err := hubClient.Resource(gvr).Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	managedCluster := &clusterv1.ManagedCluster{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		uManagedCluster.UnstructuredContent(),
		managedCluster,
	)
	if err != nil {
		return nil, err
	}

	return managedCluster, nil
}
