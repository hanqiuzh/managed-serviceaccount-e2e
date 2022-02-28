package utils

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

var gvrMCH = schema.GroupVersionResource{
	Group:    "operator.open-cluster-management.io",
	Version:  "v1",
	Resource: "multiclusterhubs",
}
var gvrMCE = schema.GroupVersionResource{
	Group:    "multicluster.openshift.io",
	Version:  "v1",
	Resource: "multiclusterengines",
}

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

// list globally and get first mch, will return no error and no obj if not found
func GetMultiClusterHub(
	hubClient dynamic.Interface,
) (*unstructured.Unstructured, error) {
	uMCHList, err := hubClient.Resource(gvrMCH).List(context.TODO(), metav1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	if errors.IsNotFound(err) || len(uMCHList.Items) < 1 {
		// ignore not found errors
		return nil, nil
	}

	return &uMCHList.Items[0], nil
}

// GetMultiClusterEngine find first MCE in the cluster, will return error if no MCE can be found
func GetMultiClusterEngine(
	hubClient dynamic.Interface,
) (*unstructured.Unstructured, error) {
	uMCEList, err := hubClient.Resource(gvrMCE).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(uMCEList.Items) < 1 {
		// not found will return error
		return nil, fmt.Errorf("MulticlusterEngine CR not found.")
	}

	return &uMCEList.Items[0], nil
}

func EnableManagedServiceAccountFeature(hubClient dynamic.Interface) error {
	name := ""
	namespace := ""
	gvr := gvrMCH
	patch := fmt.Sprintf(
		`[{"op":"%s","path":"%s","value":%s}]`,
		"add", "/spec/componentConfig", `{"managedServiceAccount":{"enable":true}}`,
	)
	// try mch
	if mch, err := GetMultiClusterHub(hubClient); err != nil {
		return err
	} else if mch == nil {
		// if not exist, try mce
		mce, err := GetMultiClusterEngine(hubClient)
		if err != nil {
			return err
		}
		name = mce.GetName()
		namespace = ""
		gvr = gvrMCE
	} else {
		name = mch.GetName()
		namespace = mch.GetNamespace()
	}

	_, err := hubClient.Resource(gvr).Namespace(namespace).Patch(context.TODO(), name, types.JSONPatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil

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