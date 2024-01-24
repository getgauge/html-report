/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	CGO_ENABLED = "CGO_ENABLED"
)

const (
	dotGauge          = ".gauge"
	plugins           = "plugins"
	goARCH            = "GOARCH"
	goOS              = "GOOS"
	x86               = "386"
	x86_64            = "amd64"
	ARM64             = "arm64"
	DARWIN            = "darwin"
	LINUX             = "linux"
	WINDOWS           = "windows"
	bin               = "bin"
	newDirPermissions = 0755
	gauge             = "gauge"
	htmlReport        = "html-report"
	deploy            = "deploy"
	pluginJSONFile    = "plugin.json"
	themesDir         = "themes"
)

var deployDir = filepath.Join(deploy, htmlReport)

func main() {
	flag.Parse()
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
			*binDir = filepath.Join(bin, fmt.Sprintf("%s_%s", platformEnv[goOS], platformEnv[goARCH]))
			fmt.Printf("Creating distro for platform => OS:%s ARCH:%s \n", platformEnv[goOS], platformEnv[goARCH])
			createDistro()
		}
	} else {
		createDistro()
	}
	fmt.Printf("Distributables created in directory => %s \n", deploy)
}

func createDistro() {
	packageName := fmt.Sprintf("%s-%s-%s.%s", htmlReport, getPluginVersion(), getGOOS(), getArch())
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
	err = os.Chdir(filepath.Join(dir, name))
	if err != nil {
		panic(fmt.Sprintf("Failed to chdir to %s: %s", filepath.Join(dir, name), err.Error()))
	}

	output, err := executeCommand("zip", "-r", filepath.Join("..", name+".zip"), ".")
	fmt.Println(output)
	if err != nil {
		panic(fmt.Sprintf("Failed to zip: %s", err.Error()))
	}
	err = os.Chdir(wd)
	if err != nil {
		panic(fmt.Sprintf("Failed to chdir to %s: %s", filepath.Join(dir, name), err.Error()))
	}
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
	fmt.Printf("Copying '%s' -> '%s'\n", src, dst)
	err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		suffix, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("Failed to find Rel(%q, %q): %v", src, path, err.Error())
		}
		return mirrorFile(path, filepath.Join(dst, suffix))
	})
	return err
}

func runProcess(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Execute %v\n", cmd.Args)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func executeCommand(command string, arg ...string) (string, error) {
	cmd := exec.Command(command, arg...)
	bytes, err := cmd.Output()
	return strings.TrimSpace(string(bytes)), err
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
		fmt.Printf("Copying %s -> %s\n", src, installDst)
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
	files[pluginJSONFile] = ""
	files[themesDir] = themesDir
	copyFiles(files, destDir)
}

func getPluginVersion() string {
	pluginProperties, err := getPluginProperties(pluginJSONFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to get properties file. %s", err.Error()))
	}
	return pluginProperties["version"].(string)
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

var (
	platformEnvs = []map[string]string{
		{goARCH: ARM64, goOS: DARWIN, CGO_ENABLED: "0"},
		{goARCH: x86_64, goOS: DARWIN, CGO_ENABLED: "0"},
		{goARCH: x86, goOS: LINUX, CGO_ENABLED: "0"},
		{goARCH: x86_64, goOS: LINUX, CGO_ENABLED: "0"},
		{goARCH: ARM64, goOS: LINUX, CGO_ENABLED: "0"},
		{goARCH: x86, goOS: WINDOWS, CGO_ENABLED: "0"},
		{goARCH: x86_64, goOS: WINDOWS, CGO_ENABLED: "0"},
	}
)

func getPluginProperties(jsonPropertiesFile string) (map[string]interface{}, error) {
	pluginPropertiesJson, err := os.ReadFile(jsonPropertiesFile)
	if err != nil {
		fmt.Printf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err.Error())
		return nil, err
	}
	var pluginJson interface{}
	if err = json.Unmarshal([]byte(pluginPropertiesJson), &pluginJson); err != nil {
		fmt.Printf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err.Error())
		return nil, err
	}
	return pluginJson.(map[string]interface{}), nil
}

func compileAcrossPlatforms() {
	for _, platformEnv := range platformEnvs {
		setEnv(platformEnv)
		fmt.Printf("Compiling for platform => OS:%s ARCH:%s \n", platformEnv[goOS], platformEnv[goARCH])
		compileGoPackage(htmlReport)
	}
}

func installPlugin(installPrefix string) {
	copyPluginFiles(deployDir)
	pluginInstallPath := filepath.Join(installPrefix, htmlReport, getPluginVersion())
	err := mirrorDir(deployDir, pluginInstallPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to mirror directory  '%s' to '%s': %s", deployDir, pluginInstallPath, err.Error()))
	}
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
	if arch == x86 {
		return "x86"
	} else if arch == ARM64 {
		return "arm64"
	}
	return "x86_64"
}

func getGOARCH() string {
	goArch := os.Getenv(goARCH)
	if goArch == "" {
		return runtime.GOARCH

	}
	return goArch
}

func getGOOS() string {
	os := os.Getenv(goOS)
	if os == "" {
		return runtime.GOOS

	}
	return os
}
