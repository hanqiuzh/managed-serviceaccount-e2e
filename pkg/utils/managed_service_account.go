package utils

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"

	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	msav1alpha1 "open-cluster-management.io/managed-serviceaccount/api/v1alpha1"
)

func unstructuredToManagedServiceAccount(
	u *unstructured.Unstructured,
) (*msav1alpha1.ManagedServiceAccount, error) {
	msa := &msav1alpha1.ManagedServiceAccount{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		u.UnstructuredContent(),
		msa,
	)
	if err != nil {
		return nil, err
	}
	return msa, nil
}

func GetManagedServiceAccount(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
	name string,
) (*msav1alpha1.ManagedServiceAccount, error) {
	gvr := schema.GroupVersionResource{
		Group:    "authentication.open-cluster-management.io",
		Version:  "v1alpha1",
		Resource: "managedserviceaccounts",
	}

	uManagedServiceAccount, err := hubClient.Resource(gvr).
		Namespace(managedCluster.Name).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	managedServiceAccount, err := unstructuredToManagedServiceAccount(uManagedServiceAccount)
	if err != nil {
		return nil, err
	}

	return managedServiceAccount, nil
}

func ListManagedServiceAccount(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
) (*msav1alpha1.ManagedServiceAccountList, error) {
	gvr := schema.GroupVersionResource{
		Group:    "authentication.open-cluster-management.io",
		Version:  "v1alpha1",
		Resource: "managedserviceaccounts",
	}

	uList, err := hubClient.Resource(gvr).
		Namespace(managedCluster.Name).
		List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	msaList, err := unstructuredListToManagedServiceAccountList(uList)
	if err != nil {
		return nil, err
	}

	return msaList, nil
}

func unstructuredListToManagedServiceAccountList(
	uList *unstructured.UnstructuredList,
) (*msav1alpha1.ManagedServiceAccountList, error) {
	msaList := &msav1alpha1.ManagedServiceAccountList{}
	for _, u := range uList.Items {
		msa, err := unstructuredToManagedServiceAccount(&u)
		if err != nil {
			return nil, err
		}
		msaList.Items = append(msaList.Items, *msa)
	}
	return msaList, nil
}

func CreateManagedServiceAccount(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
	namePrefix string,
) (*msav1alpha1.ManagedServiceAccount, error) {
	gvr := schema.GroupVersionResource{
		Group:    "authentication.open-cluster-management.io",
		Version:  "v1alpha1",
		Resource: "managedserviceaccounts",
	}

	newManagedServiceAccount := &msav1alpha1.ManagedServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ManagedServiceAccount",
			APIVersion: "authentication.open-cluster-management.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: namePrefix,
			Namespace:    managedCluster.Name,
		},
		Spec: msav1alpha1.ManagedServiceAccountSpec{
			Rotation: msav1alpha1.ManagedServiceAccountRotation{
				Enabled: true,
				Validity: metav1.Duration{
					Duration: time.Hour,
				},
			},
		},
	}

	uNewManagedServiceAccount, err := toUnstructured(newManagedServiceAccount)
	if err != nil {
		return nil, err
	}

	uManagedServiceAccount, err := hubClient.
		Resource(gvr).
		Namespace(managedCluster.Name).
		Create(
			context.TODO(),
			uNewManagedServiceAccount,
			metav1.CreateOptions{},
		)

	if err != nil {
		return nil, err
	}

	managedServiceAccount, err := unstructuredToManagedServiceAccount(uManagedServiceAccount)
	if err != nil {
		return nil, err
	}

	return managedServiceAccount, nil
}

func DoesManagedServiceAccountExist(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
	name string,
) bool {
	_, err := GetManagedServiceAccount(hubClient, managedCluster, name)
	if errors.IsNotFound(err) {
		return false
	}
	// NOTE: only false is trustworthy true is not
	return true
}

func IsManagedServiceAccountComplete(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
	name string,
) bool {
	tokenReported := false
	secretCreated := false

	managedServiceAccount, err := GetManagedServiceAccount(hubClient, managedCluster, name)
	if err != nil {
		klog.Errorf("error %v", err)
		return false
	}

	for _, condition := range managedServiceAccount.Status.Conditions {
		switch condition.Type {
		case msav1alpha1.ConditionTypeSecretCreated:
			if condition.Status == metav1.ConditionTrue {
				secretCreated = true
			}
		case msav1alpha1.ConditionTypeTokenReported:
			if condition.Status == metav1.ConditionTrue {
				tokenReported = true
			}
		}
	}

	// klog.Info(secretCreated, tokenReported)
	return secretCreated && tokenReported
}

func unstructuredToSecret(
	u *unstructured.Unstructured,
) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func getSecret(
	hubClient dynamic.Interface,
	name string,
	namespace string,
) (*corev1.Secret, error) {
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	uSecret, err := hubClient.
		Resource(gvr).
		Namespace(namespace).
		Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	secret, err := unstructuredToSecret(uSecret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func GetManagedServiceAccountSecret(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
	name string,
) (*corev1.Secret, error) {
	managedServiceAccount, err := GetManagedServiceAccount(hubClient, managedCluster, name)
	if err != nil {
		return nil, err
	}

	secretName := managedServiceAccount.Status.TokenSecretRef.Name

	secret, err := getSecret(hubClient, secretName, managedCluster.Name)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func GetManagedServiceAccountToken(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
	name string,
) (string, error) {
	secret, err := GetManagedServiceAccountSecret(hubClient, managedCluster, name)
	if err != nil {
		return "", err
	}

	if len(secret.Data["token"]) == 0 {
		return "", fmt.Errorf("empty token")
	}

	return string(secret.Data["token"]), nil
}

func DeleteManagedServiceAccount(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
	name string,
) error {
	gvr := schema.GroupVersionResource{
		Group:    "authentication.open-cluster-management.io",
		Version:  "v1alpha1",
		Resource: "managedserviceaccounts",
	}

	err := hubClient.Resource(gvr).Namespace(managedCluster.Name).Delete(
		context.TODO(),
		name,
		metav1.DeleteOptions{},
	)

	return err
}

func GetManagedServiceAccountUserName(
	hubClient dynamic.Interface,
	managedCluster *clusterv1.ManagedCluster,
	managedServiceAccountName string,
) (string, error) {
	managedServiceAccountAddon, err := GetManagedServiceAccountAddon(hubClient, managedCluster)
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("system:serviceaccount:%s:%s", managedServiceAccountAddon.Spec.InstallNamespace, managedServiceAccountName)

	return name, nil
}

func ValidateManagedServiceAccountToken(
	mcDynClient dynamic.Interface,
	token string,
	expectedUserName string,
) (bool, error) {
	gvr := schema.GroupVersionResource{
		Group:    "authentication.k8s.io",
		Version:  "v1",
		Resource: "tokenreviews",
	}

	newTokenReview := &authv1.TokenReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TokenReview",
			APIVersion: "authentication.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "token-review-request",
		},
		Spec: authv1.TokenReviewSpec{
			Token: token,
		},
	}

	uNewTokenReview, err := toUnstructured(newTokenReview)
	if err != nil {
		return false, err
	}

	uCreatedTokenReview, err := mcDynClient.Resource(gvr).Create(context.TODO(), uNewTokenReview, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	createdTokenReview, err := unstructuredToTokenReview(uCreatedTokenReview)
	if err != nil {
		return false, err
	}

	if createdTokenReview.Status.Authenticated == false {
		return false, fmt.Errorf("fail to authenticate")
	}

	if createdTokenReview.Status.User.Username != expectedUserName {
		return false, fmt.Errorf("username does not match %s", expectedUserName)
	}

	return true, nil
}

func unstructuredToTokenReview(u *unstructured.Unstructured) (*authv1.TokenReview, error) {
	tr := &authv1.TokenReview{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		u.UnstructuredContent(),
		tr,
	)
	if err != nil {
		return nil, err
	}
	return tr, nil
}
