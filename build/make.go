// Copyright 2015 ThoughtWorks, Inc.

// This file is part of getgauge/html-report.

// getgauge/html-report is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// getgauge/html-report is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with getgauge/html-report.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/getgauge/common"
)

const (
	CGO_ENABLED = "CGO_ENABLED"
)

const (
	dotGauge          = ".gauge"
	plugins           = "plugins"
	GOARCH            = "GOARCH"
	GOOS              = "GOOS"
	X86               = "386"
	X86_64            = "amd64"
	DARWIN            = "darwin"
	LINUX             = "linux"
	WINDOWS           = "windows"
	bin               = "bin"
	newDirPermissions = 0755
	gauge             = "gauge"
	htmlReport        = "html-report"
	deploy            = "deploy"
	pluginJsonFile    = "plugin.json"
	reportTemplate    = "report-template"
	GAUGE_MESSAGES    = "gauge_messages"
)

var deployDir = filepath.Join(deploy, htmlReport)
var buildMetadata string

func main() {
	flag.Parse()
	if *nightly {
		buildMetadata = fmt.Sprintf("nightly-%s", time.Now().Format(common.NightlyDatelayout))
	}

	if *install {
		updatePluginInstallPrefix()
		installPlugin(*pluginInstallPrefix)
	} else if *distro {
		createPluginDistro(*allPlatforms)
	} else {
		compile()
	}
}

func compile() {
	if *allPlatforms {
		compileAcrossPlatforms()
	} else {
		compileGoPackage(htmlReport)
	}
}

func createPluginDistro(forAllPlatforms bool) {
	if forAllPlatforms {
		for _, platformEnv := range platformEnvs {
			setEnv(platformEnv)
			*binDir = filepath.Join(bin, fmt.Sprintf("%s_%s", platformEnv[GOOS], platformEnv[GOARCH]))
			fmt.Printf("Creating distro for platform => OS:%s ARCH:%s \n", platformEnv[GOOS], platformEnv[GOARCH])
			createDistro()
		}
	} else {
		createDistro()
	}
	log.Printf("Distributables created in directory => %s \n", deploy)
}

func createDistro() {
	packageName := fmt.Sprintf("%s-%s-%s.%s", htmlReport, getPluginVersionWithBuildInfo(), getGOOS(), getArch())
	distroDir := filepath.Join(deploy, packageName)
	copyPluginFiles(distroDir)
	createZipFromUtil(deploy, packageName)
	os.RemoveAll(distroDir)
}

func createZipFromUtil(dir, name string) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Chdir(filepath.Join(dir, name))
	output, err := executeCommand("zip", "-r", filepath.Join("..", name+".zip"), ".")
	fmt.Println(output)
	if err != nil {
		panic(fmt.Sprintf("Failed to zip: %s", err))
	}
	os.Chdir(wd)
}

func isExecMode(mode os.FileMode) bool {
	return (mode & 0111) != 0
}

func mirrorFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if sfi.Mode()&os.ModeType != 0 {
		log.Fatalf("mirrorFile can't deal with non-regular file %s", src)
	}
	dfi, err := os.Stat(dst)
	if err == nil &&
		isExecMode(sfi.Mode()) == isExecMode(dfi.Mode()) &&
		(dfi.Mode()&os.ModeType == 0) &&
		dfi.Size() == sfi.Size() &&
		dfi.ModTime().Unix() == sfi.ModTime().Unix() {
		// Seems to not be modified.
		return nil
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, newDirPermissions); err != nil {
		return err
	}

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	n, err := io.Copy(df, sf)
	if err == nil && n != sfi.Size() {
		err = fmt.Errorf("copied wrong size for %s -> %s: copied %d; want %d", src, dst, n, sfi.Size())
	}
	cerr := df.Close()
	if err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(dst, sfi.Mode())
	}
	if err == nil {
		err = os.Chtimes(dst, sfi.ModTime(), sfi.ModTime())
	}
	return err
}

func mirrorDir(src, dst string) error {
	log.Printf("Copying '%s' -> '%s'\n", src, dst)
	err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		suffix, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("Failed to find Rel(%q, %q): %v", src, path, err)
		}
		return mirrorFile(path, filepath.Join(dst, suffix))
	})
	return err
}

func set(envName, envValue string) {
	log.Printf("%s = %s\n", envName, envValue)
	err := os.Setenv(envName, envValue)
	if err != nil {
		panic(err)
	}
}

func runProcess(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Execute %v\n", cmd.Args)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func executeCommand(command string, arg ...string) (string, error) {
	cmd := exec.Command(command, arg...)
	bytes, err := cmd.Output()
	return strings.TrimSpace(fmt.Sprintf("%s", bytes)), err
}

func compileGoPackage(packageName string) {
	runProcess("go", "build", "-o", getGaugeExecutablePath(htmlReport))
}

func getGaugeExecutablePath(file string) string {
	return filepath.Join(getBinDir(), getExecutableName(file))
}

func getExecutableName(file string) string {
	if getGOOS() == "windows" {
		return file + ".exe"
	}
	return file
}

func getBinDir() string {
	if *binDir != "" {
		return *binDir
	}
	return filepath.Join(bin, fmt.Sprintf("%s_%s", getGOOS(), getGOARCH()))
}

// key will be the source file and value will be the target
func copyFiles(files map[string]string, installDir string) {
	for src, dst := range files {
		base := filepath.Base(src)
		installDst := filepath.Join(installDir, dst)
		log.Printf("Copying %s -> %s\n", src, installDst)
		stat, err := os.Stat(src)
		if err != nil {
			panic(err)
		}
		if stat.IsDir() {
			err = mirrorDir(src, installDst)
		} else {
			err = mirrorFile(src, filepath.Join(installDst, base))
		}
		if err != nil {
			panic(err)
		}
	}
}

func copyPluginFiles(destDir string) {
	files := make(map[string]string)
	if getGOOS() == "windows" {
		files[filepath.Join(getBinDir(), htmlReport+".exe")] = bin
	} else {
		files[filepath.Join(getBinDir(), htmlReport)] = bin
	}
	files[pluginJsonFile] = ""
	files[reportTemplate] = reportTemplate
	copyFiles(files, destDir)
}

func getPluginVersion() string {
	pluginProperties, err := getPluginProperties(pluginJsonFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to get properties file. %s", err))
	}
	return pluginProperties["version"].(string)
}

func getPluginVersionWithBuildInfo() string {
	version := getPluginVersion()
	if buildMetadata != "" {
		version += fmt.Sprintf(".%s", buildMetadata)
	}
	return version

}

func moveOSBinaryToCurrentOSArchDirectory(targetName string) {
	destDir := path.Join(bin, fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
	moveBinaryToDirectory(path.Base(targetName), destDir)
}

func moveBinaryToDirectory(target, destDir string) error {
	if runtime.GOOS == "windows" {
		target = target + ".exe"
	}
	srcFile := path.Join(bin, target)
	destFile := path.Join(destDir, target)
	if err := os.MkdirAll(destDir, newDirPermissions); err != nil {
		return err
	}
	if err := mirrorFile(srcFile, destFile); err != nil {
		return err
	}
	return os.Remove(srcFile)
}

func setEnv(envVariables map[string]string) {
	for k, v := range envVariables {
		os.Setenv(k, v)
	}
}

var install = flag.Bool("install", false, "Install to the specified prefix")
var pluginInstallPrefix = flag.String("plugin-prefix", "", "Specifies the prefix where the plugin will be installed")
var distro = flag.Bool("distro", false, "Creates distributables for the plugin")
var allPlatforms = flag.Bool("all-platforms", false, "Compiles or creates distributables for all platforms windows, linux, darwin both x86 and x86_64")
var binDir = flag.String("bin-dir", "", "Specifies OS_PLATFORM specific binaries to install when cross compiling")
var nightly = flag.Bool("nightly", false, "Adds nightly build information")

var (
	platformEnvs = []map[string]string{
		map[string]string{GOARCH: X86, GOOS: DARWIN, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: DARWIN, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: LINUX, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: LINUX, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86, GOOS: WINDOWS, CGO_ENABLED: "0"},
		map[string]string{GOARCH: X86_64, GOOS: WINDOWS, CGO_ENABLED: "0"},
	}
)

func getPluginProperties(jsonPropertiesFile string) (map[string]interface{}, error) {
	pluginPropertiesJson, err := ioutil.ReadFile(jsonPropertiesFile)
	if err != nil {
		fmt.Printf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err)
		return nil, err
	}
	var pluginJson interface{}
	if err = json.Unmarshal([]byte(pluginPropertiesJson), &pluginJson); err != nil {
		fmt.Printf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err)
		return nil, err
	}
	return pluginJson.(map[string]interface{}), nil
}

func compileAcrossPlatforms() {
	for _, platformEnv := range platformEnvs {
		setEnv(platformEnv)
		fmt.Printf("Compiling for platform => OS:%s ARCH:%s \n", platformEnv[GOOS], platformEnv[GOARCH])
		compileGoPackage(htmlReport)
	}
}

func installPlugin(installPrefix string) {
	copyPluginFiles(deployDir)
	pluginInstallPath := filepath.Join(installPrefix, htmlReport, getPluginVersionWithBuildInfo())
	mirrorDir(deployDir, pluginInstallPath)
}

func updatePluginInstallPrefix() {
	if *pluginInstallPrefix == "" {
		if runtime.GOOS == "windows" {
			*pluginInstallPrefix = os.Getenv("APPDATA")
			if *pluginInstallPrefix == "" {
				panic(fmt.Errorf("Failed to find AppData directory"))
			}
			*pluginInstallPrefix = filepath.Join(*pluginInstallPrefix, gauge, plugins)
		} else {
			userHome := getUserHome()
			if userHome == "" {
				panic(fmt.Errorf("Failed to find User Home directory"))
			}
			*pluginInstallPrefix = filepath.Join(userHome, dotGauge, plugins)
		}
	}
}

func getUserHome() string {
	return os.Getenv("HOME")
}

func getArch() string {
	arch := getGOARCH()
	if arch == X86 {
		return "x86"
	}
	return "x86_64"
}

func getGOARCH() string {
	goArch := os.Getenv(GOARCH)
	if goArch == "" {
		return runtime.GOARCH

	}
	return goArch
}

func getGOOS() string {
	os := os.Getenv(GOOS)
	if os == "" {
		return runtime.GOOS

	}
	return os
}
