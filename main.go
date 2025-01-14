package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"zertuserver/internal/app/servers"
	"zertuserver/internal/app/services"
	"zertuserver/internal/pkg/storage"
	"zertuserver/pkg/config"
	"zertuserver/pkg/logger"
)

var (
	VersionInfo string
	BuildInfo   string
)

func main() {
	var runMode string
	if len(os.Args) > 1 {
		runMode = os.Args[1]
	} else {
		runMode = "dev"
	}
	// 打印运行信息，logger未初始化，使用fmt
	fmt.Println("------------------------------------------------")
	fmt.Printf("     build at: %s\n     version: %s\n     runMode: %s\n", BuildInfo, VersionInfo, runMode)
	fmt.Println("------------------------------------------------")

	// 依次初始化各模块 日志 配置文件 数据库 业务
	logger.Init()
	config.Init(runMode)
	storage.Init()
	services.Init()
	err2 := services.RtuService().Start()
	if nil != err2 {
		panic(err2)
	}
	go func() {
		err := servers.ApiServer().Start()
		if nil != err && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	sCh := make(chan os.Signal)
	signal.Notify(sCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-sCh
	doQuit()
}

func doQuit() {
	_ = servers.ApiServer().Stop()
	_ = services.RtuService().Stop()
}
