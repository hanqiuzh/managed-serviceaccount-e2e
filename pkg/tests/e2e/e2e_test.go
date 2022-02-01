package base_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	libgocmd "github.com/stolostron/library-e2e-go/pkg/cmd"
	libgooptions "github.com/stolostron/library-e2e-go/pkg/options"
	"github.com/stolostron/managed-serviceaccount-e2e/pkg/clients"
	"github.com/stolostron/managed-serviceaccount-e2e/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

var _ = Describe("e2e", Ordered, func() {
	var hubClient dynamic.Interface
	var mcClient dynamic.Interface
	var managedCluster *clusterv1.ManagedCluster
	var managedServiceAccountName string

	BeforeAll(func() {
		//initialize options
		err := libgooptions.LoadOptions(libgocmd.End2End.OptionsFile)
		Expect(err).To(BeNil())

		//initialize hub dynamic client
		hubClient, err = clients.GetHubDynamicClient()
		Expect(err).Should(BeNil())

		//find a managed cluster to do the test on
		managedCluster, err = utils.GetImportedCluster(hubClient)
		Expect(err).Should(BeNil())

		//initialize managedcluster dynamic client
		mcClient, err = clients.GetManagedClusterDynamicClient(managedCluster.Name)
		Expect(err).Should(BeNil())
	})

	It("[P1][Sev1][server-foundation] able to enable managed-serviceaccount addon", func() {
		// check if managed-serviceaccount addon is already enabled
		managedServiceAccountAddon, err := utils.GetManagedServiceAccountAddon(hubClient, managedCluster)
		// skip this test if already enabled
		if err != nil {
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		}

		if managedServiceAccountAddon != nil {
			Skip("ManagedServiceAccount addon already enabled")
		}

		//create managed-serviceaccount addon
		By("Creating ManagedClusterAddon in ManagedCluster namespace")
		managedServiceAccountAddon, err = utils.CreateManagedServiceAccountAddon(
			hubClient,
			managedCluster,
		)
		Expect(err).Should(BeNil())
		Expect(managedServiceAccountAddon).NotTo(BeNil())

		//eventually managed-serviceaccount addon should be availble
		Eventually(func() bool {
			return utils.IsManagedServiceAccountAddonAvailable(hubClient, managedCluster)
		}, time.Minute*5, time.Second*10).Should(BeTrue())
	})

	It("[P1][Sev1][server-foundation] able to create managed-serviceaccount", func() {
		By("creating a ManagedServiceAccount in ManagedCluster namespace")
		//create managed serviceaccount
		createdManagedServiceAccount, err := utils.CreateManagedServiceAccount(
			hubClient,
			managedCluster,
			"e2e-",
		)
		Expect(err).Should(BeNil())
		Expect(createdManagedServiceAccount).ShouldNot(BeNil())

		//eventually managed serviceaccount status condition should contain
		// - "TokenReported"
		// - "SecretCreated"
		Eventually(func() bool {
			return utils.IsManagedServiceAccountComplete(
				hubClient,
				managedCluster,
				createdManagedServiceAccount.Name,
			)
		}, time.Minute*1, time.Second*10).Should(BeTrue())

		managedServiceAccountName = createdManagedServiceAccount.Name
	})

	It("[P1][Sev1][server-foundation] managed serviceaccount should generated valid token secret", func() {
		token, err := utils.GetManagedServiceAccountToken(
			hubClient,
			managedCluster,
			managedServiceAccountName,
		)
		Expect(err).Should(BeNil())
		Expect(token).ShouldNot(BeEmpty())

		username, err := utils.GetManagedServiceAccountUserName(
			hubClient,
			managedCluster,
			managedServiceAccountName,
		)
		Expect(err).Should(BeNil())
		Expect(username).ShouldNot(BeEmpty())

		Expect(utils.ValidateManagedServiceAccountToken(mcClient, token, username)).Should(BeTrue())
	})

	It("[P1][Sev1][server-foundation] able to delete managed-serviceaccount", func() {
		//managed-serviceaccount addon shouldnt already be installed
		managedServiceAccount, err := utils.GetManagedServiceAccount(hubClient, managedCluster, managedServiceAccountName)
		Expect(err).Should(BeNil())
		Expect(managedServiceAccount).NotTo(BeNil())

		//install managed-serviceaccount addon
		err = utils.DeleteManagedServiceAccount(hubClient, managedCluster, managedServiceAccountName)
		Expect(err).Should(BeNil())

		//eventually managed-serviceaccount addon to be deleted
		Eventually(func() bool {
			return utils.DoesManagedServiceAccountExist(hubClient, managedCluster, managedServiceAccountName)
		}, time.Minute*5, time.Second*10).Should(BeFalse())
	})

	It("[P1][Sev1][server-foundation] able to disable managed-serviceaccount addon", func() {
		//managed-serviceaccount addon shouldnt already be installed
		managedServiceAccountAddon, err := utils.GetManagedServiceAccountAddon(hubClient, managedCluster)
		Expect(err).Should(BeNil())
		Expect(managedServiceAccountAddon).NotTo(BeNil())

		//install managed-serviceaccount addon
		err = utils.DeleteManagedServiceAccountAddon(hubClient, managedCluster)
		Expect(err).Should(BeNil())

		//eventually managed-serviceaccount addon to be deleted
		Eventually(func() bool {
			return utils.DoesManagedServiceAccountAddonExist(hubClient, managedCluster)
		}, time.Minute*5, time.Second*10).Should(BeFalse())
	})
})
