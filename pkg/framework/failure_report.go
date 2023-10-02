package framework

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"github.com/redhat-appstudio/e2e-tests/pkg/logs"
)

var namespaces = map[string]string{
	"Build Service":       "build-service",
	"JVM Build Service":   "jvm-build-service",
	"Application Service": "application-service",
	"Image Controller":    "image-controller",
}

func ReportFailure() func() {
	return func() {
		report := CurrentSpecReport()
		if report.Failed() {
			msg := ""
			now := time.Now()
			msg += strings.Repeat("*", 20)
			msg += "\nTest started at " + report.StartTime.String() + "\nTest ended at " + now.String()
			msg += fmt.Sprintf("\nControllers logs are stored here: %s\n", getControllersLogsLocation())
			if os.Getenv("CI") == "true" {
				msg += fmt.Sprintf("\nRHTAP custom resources are archived here: %s\n", GetGeneralArtifactsLocation()+"/redhat-appstudio-gather/artifacts/")
				msg += fmt.Sprintf("\nLogs from all OpenShift pods are archived here: %s\n", GetGeneralArtifactsLocation()+"/redhat-appstudio-hypershift-gather/artifacts/pods/")
			}
			msg += strings.Repeat("*", 20)
			AddReportEntry("DEBUG", msg)
		}
	}
}

func getControllersLogsLocation() (path string) {
	return GetFinalArtifactsLocation() + "/rhtap-controllers-logs"
}

func getLocalControllerLogsLocation() string {
	return GetArtifactsDir() + "/rhtap-controllers-logs"
}

func StoreControllersLogs(ki kubernetes.Interface) {
	logsMap := map[string][]byte{}
	logsDir := getLocalControllerLogsLocation()

	if os.Getenv("CI") == "true" {
		logsDir = os.Getenv("ARTIFACT_DIR")
	}
	for _, v := range namespaces {
		logs := ""
		podInterface := ki.CoreV1().Pods(v)
		pods, err := podInterface.List(context.Background(), metav1.ListOptions{})

		if err != nil {
			logs += "Error listing pods: " + err.Error() + "\n"
		} else {
			for _, pod := range pods.Items {
				containers := []corev1.Container{}
				containers = append(containers, pod.Spec.InitContainers...)
				containers = append(containers, pod.Spec.Containers...)
				for _, container := range containers {
					req := podInterface.GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name})
					log, err := innerDumpPod(req, container.Name)
					if err != nil {
						log += "Error getting logs: " + err.Error() + "\n"
					}
					logs += log
				}
			}
		}
		logsMap[v+".log"] = []byte(logs)
	}
	if err := logs.StoreArtifactsToDir(logsMap, logsDir+"/rhtap-controllers-logs"); err != nil {
		klog.Errorf("error storing artifacts: %+v", err)
	}
}

func ReportFailures(ki kubernetes.Interface) func() {
	return func() {
		report := CurrentSpecReport()
		if report.Failed() {
			now := time.Now()
			AddReportEntry("timing", "Test started at "+report.StartTime.String()+
				"\nTest ended at "+now.String())
			if ki == nil {
				return
			}
			for k, v := range namespaces {
				msg := "\n========= " + k + " =========\n\n"
				podInterface := ki.CoreV1().Pods(v)
				pods, err := podInterface.List(context.Background(), metav1.ListOptions{})
				if err != nil {
					msg += "Error listing pods: " + err.Error() + "\n"
				} else {
					for _, pod := range pods.Items {
						containers := []corev1.Container{}
						containers = append(containers, pod.Spec.InitContainers...)
						containers = append(containers, pod.Spec.Containers...)
						for _, container := range containers {
							req := podInterface.GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name})
							logs, err := innerDumpPod(req, container.Name)
							if err != nil {
								msg += "Error getting logs: " + err.Error() + "\n"
							} else {
								msg += FilterLogs(logs, report.StartTime) + "\n"
							}
						}
					}
				}
				AddReportEntry(v, msg)
				//logs.StoreArtifacts()
			}
		}
	}
}

func FilterLogs(logs string, start time.Time) string {

	//bit of a hack, the logs are in different formats and are not always valid JSON
	//just look for RFC 3339 dates line by line, once we find one after the start time dump the
	//rest of the lines
	lines := strings.Split(logs, "\n")
	ret := []string{}
	rfc3339Pattern := `(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2}))`

	re := regexp.MustCompile(rfc3339Pattern)
	for pos, i := range lines {
		match := re.FindStringSubmatch(i)

		if match != nil {
			dateString := match[1]
			ts, err := time.Parse(time.RFC3339, dateString)
			if err != nil {
				ret = append(ret, "Invalid Time, unable to parse date: "+i)
			} else if ts.Equal(start) || ts.After(start) {
				ret = append(ret, lines[pos:]...)
				break
			}
		}
	}

	return strings.Join(ret, "\n")

}

func innerDumpPod(req *rest.Request, containerName string) (string, error) {
	var readCloser io.ReadCloser
	var err error
	readCloser, err = req.Stream(context.TODO())
	if err != nil {
		print(fmt.Sprintf("error getting pod logs for container %s: %s", containerName, err.Error()))
		return "", err
	}
	defer func(readCloser io.ReadCloser) {
		err := readCloser.Close()
		if err != nil {
			print(fmt.Sprintf("Failed to close ReadCloser reading pod logs for container %s: %s", containerName, err.Error()))
		}
	}(readCloser)
	var b []byte
	b, err = io.ReadAll(readCloser)
	return string(b), err
}
