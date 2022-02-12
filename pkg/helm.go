package pkg

import (
	"errors"
	"fmt"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	"log"
	"net/http"
	"os"
	"path"
)

const (
	VelaCoreChartURLTemp = "https://kubevelacharts.oss-cn-hangzhou.aliyuncs.com/core/vela-core-%s.tgz"
	DefaultSemver        = "1.2.3"
)

var (
	CachePath      = path.Join(VelaDir(), "cache")
	CacheChartTemp = func() string { return path.Join(CachePath, "vela-core-%s.tgz") }()
)

func init() {
	err := os.MkdirAll(CachePath, 0755)
	if err != nil {
		panic(err)
	}
}

func checkChartExist(semver string) bool {
	filePath := fmt.Sprintf(CacheChartTemp, semver)
	_, err := os.Stat(filePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	} else if err != nil {
		klog.ErrorS(err, "fail to stat chart file")
	}
	return true
}

// prepareChart will return cache chart to local and return its path
func prepareChart(semver string) (string, error) {
	if checkChartExist(semver) {
		return chartCachePathForSemver(semver), nil
	}

	// download the chart
	if semver == "" {
		semver = DefaultSemver
	}
	chartURL := fmt.Sprintf(VelaCoreChartURLTemp, semver)
	resp, err := http.Get(chartURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	newFile, err := os.Create(chartCachePathForSemver(semver))
	if err != nil {
		klog.ErrorS(err, "fail to create chart cache file")
		return "", err
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, resp.Body)
	if err != nil {
		klog.ErrorS(err, "fail to copy chart file content")
		return "", err
	}
	return chartCachePathForSemver(semver), nil
}

func InstallVelaCore(opts HelmOpts) error {
	klog.Info("Installing KubeVela Helm chart, please hold...")
	chartPath, err := prepareChart(opts.Version)
	if err != nil {
		klog.ErrorS(err, "fail to prepare vela-core chart")
	}
	klog.Infof("Successfully prepare chart file in %s", chartPath)
	chart, err := loader.Load(chartPath)
	if err != nil {
		panic(err)
	}

	releaseName := "kubevela"
	releaseNamespace := "vela-system"

	actionConfig := new(action.Configuration)
	var settings = cli.New()
	settings.SetNamespace(releaseNamespace)
	helmDriver := os.Getenv("HELM_DRIVER")
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), helmDriver, debug); err != nil {
		log.Fatal(err)
	}

	iCli := action.NewInstall(actionConfig)
	iCli.Namespace = releaseNamespace
	iCli.ReleaseName = releaseName
	iCli.CreateNamespace = true
	_, err = iCli.Run(chart, nil)
	if err != nil {
		klog.ErrorS(err, "fail to run helm install action")
		return err
	}
	return nil
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	klog.Infof(format, v...)
}

func chartCachePathForSemver(semver string) string {
	return path.Join(VelaDir(), fmt.Sprintf("vela-core-%s.tgz", semver))
}
