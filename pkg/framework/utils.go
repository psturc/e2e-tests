package framework

import (
	"fmt"
	"os"
)

func (c *ControllerHub) StoreAllArtifactsForNamespace(namespace string) error {
	var finalError string
	finalError = appendErrorToString(finalError, c.HasController.StoreAllApplications(namespace))
	finalError = appendErrorToString(finalError, c.HasController.StoreAllComponents(namespace))
	finalError = appendErrorToString(finalError, c.HasController.StoreAllComponentDetectionQueries(namespace))
	finalError = appendErrorToString(finalError, c.IntegrationController.StoreAllSnapshots(namespace))
	finalError = appendErrorToString(finalError, c.TektonController.StoreAllPipelineRuns(namespace))
	finalError = appendErrorToString(finalError, c.CommonController.StoreAllPods(namespace))
	finalError = appendErrorToString(finalError, c.CommonController.StoreAllSnapshotEnvironmentBindings(namespace))
	finalError = appendErrorToString(finalError, c.GitOpsController.StoreAllDeploymentTargetClaims(namespace))
	finalError = appendErrorToString(finalError, c.GitOpsController.StoreAllDeploymentTargetClasses(namespace))
	finalError = appendErrorToString(finalError, c.GitOpsController.StoreAllDeploymentTargets(namespace))
	finalError = appendErrorToString(finalError, c.GitOpsController.StoreAllEnvironments(namespace))
	finalError = appendErrorToString(finalError, c.GitOpsController.StoreAllGitOpsDeployments(namespace))
	if len(finalError) > 0 {
		return fmt.Errorf(finalError)
	}
	return nil
}

func appendErrorToString(baseString string, err error) string {
	if err != nil {
		return fmt.Sprintf("%s\n%s", baseString, err)
	}
	return baseString
}

func GetArtifactsDir() string {
	if os.Getenv("CI") == "true" {
		return os.Getenv("ARTIFACT_DIR")
	}
	return getLocalArtifactsDir()
}

func GetFinalArtifactsLocation() string {
	if os.Getenv("CI") == "true" {
		var path string
		path += "https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/origin-ci-test"
		if os.Getenv("PULL_NUMBER") != "" {
			path += fmt.Sprintf("/pr-logs/pull/%s_%s/%s", os.Getenv("REPO_OWNER"), os.Getenv("REPO_NAME"), os.Getenv("PULL_NUMBER"))
		}
		path += fmt.Sprintf("/%s/%s/artifacts/redhat-appstudio-e2e/redhat-appstudio-e2e/artifacts", os.Getenv("JOB_NAME"), os.Getenv("BUILD_ID"))
		return path
	}
	return getLocalArtifactsDir()
}

func getLocalArtifactsDir() string {
	wd, _ := os.Getwd()
	artifactsDir := fmt.Sprintf("%s/tmp/artifacts", wd)
	os.MkdirAll(artifactsDir, 0775)
	return artifactsDir
}
