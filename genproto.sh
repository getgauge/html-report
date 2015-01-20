#!/bin/sh

#Using protoc version 2.5.0

cd gauge-proto
PATH=$PATH:$GOPATH/bin protoc --go_out=../gauge_messages spec.proto
PATH=$PATH:$GOPATH/bin protoc --go_out=../gauge_messages messages.proto

cd ..
sed  -i.backup '/import gauge_messages1 "spec.pb"/d' gauge_messages/messages.pb.go && sed  -i.backup 's/gauge_messages1.//g' gauge_messages/messages.pb.go && rm gauge_messages/messages.pb.go.backup
go fmt github.com/getgauge/html-report/...
