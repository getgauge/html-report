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
	"log"
	"net"
	"os"
	"time"

	"github.com/getgauge/common"

	"github.com/getgauge/html-report/env"

	"github.com/getgauge/html-report/gauge_messages"
	"github.com/getgauge/html-report/logger"
	"github.com/golang/protobuf/proto"
)

const pluginID = "html-report"

type GaugeResultHandlerFn func(*gauge_messages.SuiteExecutionResult)
type GaugeResultItemHandlerFn func(*gauge_messages.SuiteExecutionResultItem)

type GaugeListener struct {
	connection          net.Conn
	onResultHandler     GaugeResultHandlerFn
	onResultItemHandler GaugeResultItemHandlerFn
	stopChan            chan bool
}

func NewGaugeListener(host string, port string, killChan chan bool) (*GaugeListener, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err == nil {
		return &GaugeListener{connection: conn, stopChan: killChan}, nil
	} else {
		return nil, err
	}
}

func (gaugeListener *GaugeListener) OnSuiteResult(h GaugeResultHandlerFn) {
	gaugeListener.onResultHandler = h
}

func (gaugeListener *GaugeListener) OnSuiteResultItem(h GaugeResultItemHandlerFn) {
	gaugeListener.onResultItemHandler = h
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
				switch message.MessageType {
				case gauge_messages.Message_KillProcessRequest:
					logger.Debug("Received Kill Message, exiting...")
					gaugeListener.connection.Close()
					os.Exit(0)
				case gauge_messages.Message_SuiteExecutionResult:
					logger.Debug("Received SuiteExecutionResult, processing...")
					go gaugeListener.sendPings()
					result := message.GetSuiteExecutionResult()
					gaugeListener.onResultHandler(result)
				case gauge_messages.Message_SuiteExecutionResultItem:
					result := message.GetSuiteExecutionResultItem()
					logger.Debug("Received SuiteExecutionResultItem for %s, processing...", result.ResultItem.FileName)
					gaugeListener.onResultItemHandler(result)
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

func (gaugeListener *GaugeListener) sendPings() {
	msg := &gauge_messages.Message{
		MessageId:   common.GetUniqueID(),
		MessageType: gauge_messages.Message_KeepAlive,
		KeepAlive:   &gauge_messages.KeepAlive{PluginId: pluginID},
	}
	m, err := proto.Marshal(msg)
	if err != nil {
		logger.Debug("Unable to marshal ping message, %s", err.Error())
		return
	}
	ping := func(b []byte, c net.Conn) {
		logger.Debug("html-report sending a keep-alive ping")
		l := proto.EncodeVarint(uint64(len(b)))
		_, err := c.Write(append(l, b...))
		if err != nil {
			logger.Debug("Unable to send ping message, %s", err.Error())
		}
	}
	ticker := time.NewTicker(interval())
	defer func() { ticker.Stop() }()

	for {
		select {
		case <-gaugeListener.stopChan:
			logger.Debug("Stopping pings")
			return
		case <-ticker.C:
			ping(m, gaugeListener.connection)
		}
	}
}

var interval = func() time.Duration {
	v := env.PluginKillTimeout()
	if v/2 < 2 {
		return 2 * time.Second
	}
	return time.Duration(v * 1000 * 1000 * 1000 / 2)
}
