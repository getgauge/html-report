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
	"net"
	"testing"
	"time"

	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/gauge_messages"
	"github.com/golang/protobuf/proto"
)

func TestPingIntervalFromPluginTimeoutConfig(t *testing.T) {
	env.PluginKillTimeout = func() int { return 30 }
	got := interval()
	if got != 15*time.Second {
		t.Errorf("expected interval to be half of timeout (=30 seconds) , got %d", got)
	}
}

func TestPingIntervalIsMinTwoSeconds(t *testing.T) {
	env.PluginKillTimeout = func() int { return 2 }
	got := interval()
	if got != 2*time.Second {
		t.Errorf("expected interval to 2 seconds , got %d", got)
	}
}

func TestSendPings(t *testing.T) {
	interval = func() time.Duration { return 100 * time.Millisecond }

	server, client := net.Pipe()
	r := make(chan bool)

	go func(c net.Conn, receive chan bool) {
		b := new(bytes.Buffer)
		data := make([]byte, 8192)
		for {
			n, err := c.Read(data)
			if err != nil {
				t.Error(err)
			}
			b.Write(data[0:n])
			messageLength, bytesRead := proto.DecodeVarint(b.Bytes())
			message := &gauge_messages.Message{}
			messageBoundary := int(messageLength) + bytesRead
			err = proto.Unmarshal(b.Bytes()[bytesRead:messageBoundary], message)
			if err != nil {
				t.Error(err)
			}
			if message.MessageType != gauge_messages.Message_KeepAlive {
				t.Errorf("Expected keepalive request, got %s", message.MessageType)
			} else {
				receive <- true
			}
		}
	}(client, r)

	l := &GaugeListener{connection: server}
	go l.sendPings()

	pings := 0
	tmr := time.NewTimer(1 * time.Second)
	for {
		select {
		case <-r:
			pings = pings + 1
		case <-tmr.C:
			if pings == 0 {
				t.Error("No ping received")
			}
			if pings < 8 {
				t.Errorf("Expected more than 8 pings, got %d", pings)
			}
			return
		}
	}
}

func TestSendPingsStopsAfterInterrupt(t *testing.T) {
	interval = func() time.Duration { return 100 * time.Millisecond }

	server, client := net.Pipe()
	c := make(chan bool)

	go func(c net.Conn, stop chan bool) {
		data := make([]byte, 8192)
		aborted := false
		for {
			select {
			case <-stop:
				aborted = true
			default:
				_, err := c.Read(data)
				if err != nil {
					t.Error(err)
				}
				if aborted {
					t.Error("received message after stop")
				}
			}
		}
	}(client, c)

	l := &GaugeListener{connection: server, stopChan: c}
	go l.sendPings()

	exit := make(chan bool)

	time.AfterFunc(1*time.Second, func() {
		c <- true
		time.AfterFunc(1*time.Second, func() { exit <- true })
	})
	for {
		if <-exit {
			return
		}
	}
}
