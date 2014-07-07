package main

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"fmt"
	"github.com/getgauge/common"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
)

const (
	reportTemplateDir = "report-template"
)

var pluginProperties map[string]interface{}
var pluginInstallationDir string
var projectRoot string

func main() {
	pluginInstallationDir = os.Getenv("plugin_root")
	if pluginInstallationDir == "" {
		fmt.Println("environment variable plugin_root is not set")
		os.Exit(1)
	}
	projectRoot = os.Getenv("project_root")
	if projectRoot == "" {
		fmt.Println("environment variable project_root is not set")
		os.Exit(1)
	}

	pluginPropertiesJson, err := ioutil.ReadFile(filepath.Join(pluginInstallationDir, "plugin.json"))
	if err != nil {
		fmt.Printf("Could not read plugin.json: %s\n", err)
		os.Exit(1)
	}
	var pluginJson interface{}
	if err = json.Unmarshal([]byte(pluginPropertiesJson), &pluginJson); err != nil {
		fmt.Printf("Could not read plugin.json: %s\n", err)
		os.Exit(1)
	}
	pluginProperties = pluginJson.(map[string]interface{})

	action := os.Getenv("html-report_action")
	if action == "execution" {
		listener, err := NewGaugeListener("localhost", os.Getenv("plugin_connection_port"))
		if err != nil {
			fmt.Println("Could not create the gauge listener")
			os.Exit(1)
		}
		listener.OnSuiteResult(createReport)
		listener.Start()
	}
}

type GaugeResultHandlerFn func(*SuiteExecutionResult)

type GaugeListener struct {
	connnection     net.Conn
	onResultHandler GaugeResultHandlerFn
}

func NewGaugeListener(host string, port string) (*GaugeListener, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err == nil {
		return &GaugeListener{connnection: conn}, nil
	} else {
		return nil, err
	}
}

func (gaugeListener *GaugeListener) OnSuiteResult(resultHandler GaugeResultHandlerFn) {
	gaugeListener.onResultHandler = resultHandler
}

func (gaugeListener *GaugeListener) Start() {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := gaugeListener.connnection.Read(data)
		if err != nil {
			if err == io.EOF {
				return
			}
		}
		buffer.Write(data[0:n])

		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			message := &Message{}
			err = proto.Unmarshal(buffer.Bytes()[bytesRead:messageLength+uint64(bytesRead)], message)
			if err != nil {
				log.Printf("[Gauge Listener] Failed to read proto message: %s\n", err.Error())
			} else {
				if *message.MessageType == Message_SuiteExecutionResult {
					result := message.GetSuiteExecutionResult()
					gaugeListener.onResultHandler(result)
					gaugeListener.connnection.Close()
					return
				}
				buffer.Reset()
			}
		}
	}
}

func createReport(suiteResult *SuiteExecutionResult) {
	contents := generateJsFileContents(suiteResult)
	reportsDir := os.Getenv(common.GaugeReportsDirEnvName)
	if reportsDir == "" {
		createDirectory(common.DefaultReportsDir)
		reportsDir = common.DefaultReportsDir
	} else {
		createDirectory(reportsDir)
	}

	currentReportDir := path.Join(projectRoot, reportsDir, "html-report")
	createDirectory(currentReportDir)
	copyReportTemplateFiles(currentReportDir)

	resultJsPath := path.Join(currentReportDir, "js", "result.js")
	err := ioutil.WriteFile(resultJsPath, contents, common.NewFilePermissions)
	if err != nil {
		fmt.Printf("Error writing file %s :%s\n", resultJsPath, err)
	}
	fmt.Printf("Sucessfully generated html reports to %s\n", currentReportDir)
}

func copyReportTemplateFiles(reportDir string) {
	pluginsDir, err := common.GetPluginsPath()
	if err != nil {
		fmt.Printf("Error finding plugins directory :%s\n", err)
		os.Exit(1)
	}
	reportTemplateDir := path.Join(pluginsDir, pluginProperties["id"].(string), pluginProperties["version"].(string), reportTemplateDir)
	err = common.MirrorDir(reportTemplateDir, reportDir)
	if err != nil {
		fmt.Printf("Error copying template directory :%s\n", err)
		os.Exit(1)
	}

}

func generateJsFileContents(suiteResult *SuiteExecutionResult) []byte {
	var buffer bytes.Buffer
	executionResultJson := marshal(suiteResult)
	itemsTypeJson := marshal(convertKeysToString(ProtoItem_ItemType_name))
	parameterTypeJson := marshal(convertKeysToString(Parameter_ParameterType_name))
	fragmentTypeJson := marshal(convertKeysToString(Fragment_FragmentType_name))

	buffer.WriteString("var gaugeExecutionResult = ")
	buffer.Write(executionResultJson)
	buffer.WriteString(";")
	buffer.WriteString("\n var itemTypesMap = ")
	buffer.Write(itemsTypeJson)
	buffer.WriteString(";")
	buffer.WriteString("\n var parameterTypesMap = ")
	buffer.Write(parameterTypeJson)
	buffer.WriteString(";")
	buffer.WriteString("\n var fragmentTypesMap = ")
	buffer.Write(fragmentTypeJson)
	buffer.WriteString(";")

	return buffer.Bytes()
}

func convertKeysToString(intKeyMap map[int32]string) map[string]string {
	stringKeyMap := make(map[string]string, 0)
	for key, val := range intKeyMap {
		stringKeyMap[fmt.Sprintf("%d", key)] = val
	}
	return stringKeyMap
}

func marshal(item interface{}) []byte {
	marshalledResult, err := json.Marshal(item)
	if err != nil {
		fmt.Printf("Failed to convert to json :%s\n", err)
		os.Exit(1)
	}
	return marshalledResult
}

func createDirectory(dir string) {
	if common.DirExists(dir) {
		return
	}
	if err := os.MkdirAll(dir, common.NewDirectoryPermissions); err != nil {
		fmt.Printf("Failed to create directory %s: %s\n", common.DefaultReportsDir, err)
		os.Exit(1)
	}
}
