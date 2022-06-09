package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/quanxiang-cloud/audit/api/restful"
	"github.com/quanxiang-cloud/audit/internal/audit"
	"github.com/quanxiang-cloud/audit/pkg/config"
	"github.com/quanxiang-cloud/audit/pkg/misc/logger"
)

var (
	configPath = flag.String("config", "../configs/config.yml", "-config 配置文件地址")
)

func main() {
	flag.Parse()

	conf, err := config.NewConfig(*configPath)
	if err != nil {
		panic(err)
	}

	err = logger.New(&conf.Log)
	if err != nil {
		panic(err)
	}

	handler, err := audit.NewHandler(conf)
	if err != nil {
		panic(err)
	}
	handler.Handler()

	// 启动路由
	router, err := restful.NewRouter(conf)
	if err != nil {
		panic(err)
	}
	go router.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			handler.Close()
			logger.Sync()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
