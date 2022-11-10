package build

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/devfile/library/pkg/util"
	"github.com/google/uuid"
	"github.com/redhat-appstudio/e2e-tests/pkg/constants"
	"github.com/redhat-appstudio/e2e-tests/pkg/utils"
	"k8s.io/klog"
	"knative.dev/pkg/apis"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-appstudio/e2e-tests/pkg/framework"
)

const (
	helloWorldComponentGitSourceRepoName = "devfile-sample-hello-world"
	pythonComponentGitSourceURL          = "https://github.com/redhat-appstudio-qe/devfile-sample-python-basic"
)

var (
	helloWorldComponentGitSourceURL = fmt.Sprintf("https://github.com/%s/%s", utils.GetEnv("GITHUB_E2E_ORGANIZATION", "redhat-appstudio-qe"), helloWorldComponentGitSourceRepoName)
)

var _ = framework.BuildSuiteDescribe("HACBS demo", Label("hacbs", "hacbs-demo"), func() {
	defer GinkgoRecover()

	f, err := framework.NewFramework()
	Expect(err).NotTo(HaveOccurred())

	var applicationName, componentName, testNamespace, outputContainerImage string
	var timeout, interval time.Duration

	Describe("HACBS-683 Build a component using HACBS pipeline", Ordered, func() {
		BeforeAll(func() {
			testNamespace = utils.GetEnv(constants.E2E_APPLICATIONS_NAMESPACE_ENV, fmt.Sprintf("hacbs-e2e-%s", util.GenerateRandomString(4)))
			_, err := f.HacbsUser.CommonController.CreateTestNamespace(testNamespace)
			Expect(err).NotTo(HaveOccurred(), "Error when creating/updating '%s' namespace: %v", testNamespace, err)

			_, err = f.HacbsUser.CommonController.CreateRegistryAuthSecret(
				"redhat-appstudio-registry-pull-secret",
				testNamespace,
				os.Getenv("QUAY_TOKEN"),
			)
			Expect(err).NotTo(HaveOccurred(), "Error when registry auth secret in namespace '%s': %v", testNamespace, err)

			applicationName = fmt.Sprintf("hacbs-demo-test-app-%s", util.GenerateRandomString(4))
			app, err := f.HacbsUser.HasController.CreateHasApplication(applicationName, testNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(utils.WaitUntil(f.HacbsUser.CommonController.ApplicationGitopsRepoExists(app.Status.Devfile), 30*time.Second)).To(
				Succeed(), fmt.Sprintf("timed out waiting for gitops content to be created for app %s in namespace %s: %+v", app.Name, app.Namespace, err),
			)

			componentName = fmt.Sprintf("%s-%s", "test-component", util.GenerateRandomString(4))

			outputContainerImage = fmt.Sprintf("quay.io/%s/test-images:%s", utils.GetQuayIOOrganization(), strings.Replace(uuid.New().String(), "-", "", -1))

			// Create a component with Git Source URL being defined
			_, err = f.HacbsUser.HasController.CreateComponent(applicationName, componentName, testNamespace, helloWorldComponentGitSourceURL, "", "", outputContainerImage, "")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("a PipelineRun is triggered after creating a component", func() {
			timeout = time.Second * 120
			interval = time.Second * 1
			Eventually(func() bool {
				pipelineRun, err := f.HacbsUser.HasController.GetComponentPipelineRun(componentName, applicationName, testNamespace, false, "")
				if err != nil {
					klog.Infoln("PipelineRun has not been created yet")
					return false
				}
				return pipelineRun.HasStarted()
			}, timeout, interval).Should(BeTrue(), "timed out when waiting for the PipelineRun to start")
		})

		It("the PipelineRun should eventually finish successfully", func() {
			timeout = time.Second * 900
			interval = time.Second * 15
			Eventually(func() bool {

				pipelineRun, err := f.HacbsUser.HasController.GetComponentPipelineRun(componentName, applicationName, testNamespace, true, "")
				Expect(err).ShouldNot(HaveOccurred())

				for _, condition := range pipelineRun.Status.Conditions {
					klog.Infof("PipelineRun %s, namespace: %s, Status.Conditions.Reason: %s\n", pipelineRun.Name, testNamespace, condition.Reason)

					if !pipelineRun.IsDone() {
						return false
					}

					if !pipelineRun.GetStatusCondition().GetCondition(apis.ConditionSucceeded).IsTrue() {
						failMessage := fmt.Sprintf("Pipelinerun '%s' didn't succeed\n", pipelineRun.Name)
						d := utils.GetFailedPipelineRunDetails(pipelineRun)
						if d.FailedContainerName != "" {
							failMessage += fmt.Sprintf("PipelineRun details:\n%+v", d)
						}
						Fail(failMessage)
					}
				}
				return true
			}, timeout, interval).Should(BeTrue(), "timed out when waiting for the PipelineRun to finish")
		})
	})
})
