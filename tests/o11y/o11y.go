package o11y

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-appstudio/e2e-tests/pkg/framework"
	"github.com/redhat-appstudio/e2e-tests/pkg/utils"
)

var _ = framework.O11ySuiteDescribe("O11Y E2E tests", Label("o11y", "HACBS"), Pending, func() {

	defer GinkgoRecover()
	var f *framework.Framework
	var err error

	Describe("O11y test", func() {
		var testNamespace string

		BeforeAll(func() {

			f, err = framework.NewFramework(O11yUser)
			Expect(err).NotTo(HaveOccurred())
			testNamespace = f.UserNamespace

		})
		AfterAll(func() {
			if !CurrentSpecReport().Failed() {
				err := f.AsKubeAdmin.CommonController.DeleteAllPipelineRunsAndTasks(testNamespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(f.SandboxController.DeleteUserSignup(f.UserName)).NotTo(BeFalse())
			}
		})

		It("E2E test to measure Egress pod by pushing images to quay", func() {

			// Get Quay Organization from ENV
			quayOrg := utils.GetEnv("DEFAULT_QUAY_ORG", "")
			Expect(quayOrg).ToNot(BeEmpty())

			// Get Quay Token from ENV
			quayToken := utils.GetEnv("QUAY_TOKEN", "")
			Expect(quayToken).ToNot(BeEmpty())

			_, err := f.AsKubeAdmin.CommonController.CreateRegistryAuthSecret(o11yUserSecret, testNamespace, quayToken)
			Expect(err).ToNot(HaveOccurred())

			err = f.AsKubeAdmin.CommonController.LinkSecretToServiceAccount(testNamespace, o11yUserSecret, O11ySA)
			Expect(err).ToNot(HaveOccurred())

			pipelineRun, err := f.AsKubeAdmin.O11yController.QuayImagePushPipelineRun(quayOrg, o11yUserSecret, testNamespace)
			Expect(err).NotTo(HaveOccurred())

			// Wait for the pipeline run to succeed
			Eventually(func() bool {
				pr, err := f.AsKubeAdmin.TektonController.GetPipelineRun(pipelineRun.Name, pipelineRun.Namespace)
				if err != nil {
					GinkgoWriter.Println("PipelineRun has not been created yet \n", err, "PR Runs:", pr)
					return false
				}
				if len(pr.Status.Conditions) > 0 {
					for _, c := range pr.Status.Conditions {
						if c.Type == "Succeeded" && c.Status == "True" {
							return true
						}
					}
				}
				return false
			}, pipelinerun_timeout, pipelinerun_interval).Should(BeTrue(), "timed out when waiting for the PipelineRun to start or complete")

			podNameRegex := ".*-buildah-quay-pod"
			query := fmt.Sprintf("last_over_time(container_network_transmit_bytes_total{namespace='%s', pod=~'%s'}[1h])", testNamespace, podNameRegex)
			result, err := f.AsKubeAdmin.O11yController.GetMetrics(query)
			Expect(err).NotTo(HaveOccurred())

			podNamesWithSize, err := f.AsKubeAdmin.O11yController.GetRegexPodNameWithSize(podNameRegex, result)
			Expect(err).NotTo(HaveOccurred())

			for podName, podSize := range podNamesWithSize {
				// Range limits are measured as part of STONEO11Y-15
				Expect(podSize).To(And(
					BeNumerically(">=", 106.00),
					BeNumerically("<=", 108.50),
				), fmt.Sprintf("%s: %.2f MB is not within the expected range.\n", podName, podSize))
			}
		})
	})
})
