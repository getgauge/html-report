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

package listener

import (
	"bytes"
	"fmt"
	"github.com/getgauge/html-report/gauge_messages"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"os"
)

type GaugeResultHandlerFn func(*gauge_messages.SuiteExecutionResult)

type GaugeListener struct {
	connection      net.Conn
	onResultHandler GaugeResultHandlerFn
}

func NewGaugeListener(host string, port string) (*GaugeListener, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err == nil {
		return &GaugeListener{connection: conn}, nil
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
		n, err := gaugeListener.connection.Read(data)
		if err != nil {
			return
		}
		buffer.Write(data[0:n])
		gaugeListener.processMessages(buffer)
	}
}

func (gaugeListener *GaugeListener) processMessages(buffer *bytes.Buffer) {
	for {
		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			message := &gauge_messages.Message{}
			messageBoundary := int(messageLength) + bytesRead
			err := proto.Unmarshal(buffer.Bytes()[bytesRead:messageBoundary], message)
			if err != nil {
				log.Printf("Failed to read proto message: %s\n", err.Error())
			} else {
				if *message.MessageType == gauge_messages.Message_KillProcessRequest {
					os.Exit(0)
				}
				if *message.MessageType == gauge_messages.Message_SuiteExecutionResult {
					result := message.GetSuiteExecutionResult()
					gaugeListener.onResultHandler(result)
					gaugeListener.connection.Close()
				}
				buffer.Next(messageBoundary)
				if buffer.Len() == 0 {
					return
				}
			}
		} else {
			return
		}
	}
}
