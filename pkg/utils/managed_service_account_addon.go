package utils

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

func unstructuredToManagedClusterAddon(
	u *unstructured.Unstructured,
) (*addonv1alpha1.ManagedClusterAddOn, error) {
	mca := &addonv1alpha1.ManagedClusterAddOn{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		u.UnstructuredContent(),
		mca,
	)
	if err != nil {
		return nil, err
	}

	return mca, nil
}

func DoesManagedServiceAccountAddonExist(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
) bool {
	_, err := GetManagedServiceAccountAddon(hubClient, managedCluster)
	if errors.IsNotFound(err) {
		return false
	}

	// NOTE: only false is trustworthy true is not
	return true
}

func IsManagedServiceAccountAddonAvailable(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
) bool {
	managedServiceAccountAddon, err := GetManagedServiceAccountAddon(hubClient, managedCluster)
	if err != nil {
		return false
	}

	for _, condition := range managedServiceAccountAddon.Status.Conditions {
		if condition.Type == addonv1alpha1.ManagedClusterAddOnConditionAvailable {
			if condition.Status == "True" {
				return true
			}
		}
	}
	return false
}

func GetManagedServiceAccountAddon(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
) (*addonv1alpha1.ManagedClusterAddOn, error) {
	gvr := schema.GroupVersionResource{
		Group:    "addon.open-cluster-management.io",
		Version:  "v1alpha1",
		Resource: "managedclusteraddons",
	}

	uManagedServiceAccountAddon, err := hubClient.Resource(gvr).Namespace(managedCluster.Name).
		Get(context.TODO(), "managed-serviceaccount", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	managedServiceAccountAddon, err := unstructuredToManagedClusterAddon(uManagedServiceAccountAddon)
	if err != nil {
		return nil, err
	}

	return managedServiceAccountAddon, nil
}

func CreateManagedServiceAccountAddon(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
) (*addonv1alpha1.ManagedClusterAddOn, error) {
	gvr := schema.GroupVersionResource{
		Group:    "addon.open-cluster-management.io",
		Version:  "v1alpha1",
		Resource: "managedclusteraddons",
	}

	managedServiceAccountAddon, err := GetManagedServiceAccountAddon(hubClient, managedCluster)
	if errors.IsNotFound(err) {
		newManagedServiceAccountAddon := &addonv1alpha1.ManagedClusterAddOn{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ManagedClusterAddOn",
				APIVersion: "addon.open-cluster-management.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "managed-serviceaccount",
				Namespace: managedCluster.Name,
			},
			Spec: addonv1alpha1.ManagedClusterAddOnSpec{
				InstallNamespace: "open-cluster-management-managed-serviceaccount",
			},
		}

		uNewManagedServiceAccountAddon, err := toUnstructured(newManagedServiceAccountAddon)
		if err != nil {
			return nil, err
		}

		uManagedServiceAccountAddon, err := hubClient.
			Resource(gvr).
			Namespace(managedCluster.Name).
			Create(
				context.TODO(),
				uNewManagedServiceAccountAddon,
				metav1.CreateOptions{},
			)

		if err != nil {
			return nil, err
		}

		managedServiceAccountAddon, err := unstructuredToManagedClusterAddon(uManagedServiceAccountAddon)
		if err != nil {
			return nil, err
		}

		return managedServiceAccountAddon, nil
	}

	return managedServiceAccountAddon, nil
}

func DeleteManagedServiceAccountAddon(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
) error {
	gvr := schema.GroupVersionResource{
		Group:    "addon.open-cluster-management.io",
		Version:  "v1alpha1",
		Resource: "managedclusteraddons",
	}

	err := hubClient.Resource(gvr).Namespace(managedCluster.Name).Delete(
		context.TODO(),
		"managed-serviceaccount",
		metav1.DeleteOptions{},
	)

	return err
}
