#!/bin/bash
sudo echo -e "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
source /etc/profile
go version

go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off
go env -w GOSUMDB="sum.golang.org"

