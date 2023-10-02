package logs

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/redhat-appstudio/e2e-tests/pkg/utils"
	"sigs.k8s.io/yaml"
)

// createArtifactDirectory creates directory for storing artifacts of current spec.
// if path to the directory is not provided, it creates it based on the ARTIFACT_DIR env var
// if none of these is provided, then it creates a "./tmp" directory
func createArtifactDirectory(dir string) (string, error) {
	if dir == "" {
		wd, _ := os.Getwd()
		artifactDir := GetEnv("ARTIFACT_DIR", fmt.Sprintf("%s/tmp", wd))
		classname := ShortenStringAddHash(CurrentSpecReport())
		dir = fmt.Sprintf("%s/%s", artifactDir, classname)
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	return dir, nil

}

// StoreResourceYaml stores yaml of given resource.
func StoreResourceYaml(resource any, name string) error {
	resourceYaml, err := yaml.Marshal(resource)
	if err != nil {
		return fmt.Errorf("error getting resource yaml: %v", err)
	}

	resources := map[string][]byte{
		name + ".yaml": resourceYaml,
	}

	return StoreArtifacts(resources)
}

// StoreArtifacts stores given artifacts under artifact directory.
func StoreArtifacts(artifacts map[string][]byte) error {
	artifactsDirectory, err := createArtifactDirectory("")
	if err != nil {
		return err
	}

	for artifact_name, artifact_value := range artifacts {
		filePath := fmt.Sprintf("%s/%s", artifactsDirectory, artifact_name)
		if err := os.WriteFile(filePath, []byte(artifact_value), 0644); err != nil {
			return err
		}
	}

	return nil
}

// StoreArtifacts stores given artifacts under artifact directory.
func StoreArtifactsToDir(artifacts map[string][]byte, path string) error {
	artifactsDirectory, err := createArtifactDirectory(path)
	if err != nil {
		return err
	}

	for artifact_name, artifact_value := range artifacts {
		filePath := fmt.Sprintf("%s/%s", artifactsDirectory, artifact_name)
		if err := os.WriteFile(filePath, []byte(artifact_value), 0644); err != nil {
			return err
		}
	}

	return nil
}
