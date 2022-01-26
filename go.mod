module github.com/stolostron/managed-serviceaccount-e2e

go 1.16

require (
	github.com/onsi/ginkgo/v2 v2.1.0
	github.com/onsi/gomega v1.18.0
	github.com/stolostron/library-e2e-go v0.0.0-20220112062416-0820a253cf3b
	github.com/stolostron/library-go v0.0.0-20220112062416-536980fdb526
	k8s.io/api v0.23.3
	k8s.io/apiextensions-apiserver v0.23.3
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v0.23.3
	k8s.io/klog v1.0.0
	open-cluster-management.io/api v0.6.0
	open-cluster-management.io/managed-serviceaccount v0.1.0
	sigs.k8s.io/controller-runtime v0.11.0
)
