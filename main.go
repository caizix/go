package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webapp_demo/dao/mysql"
	"webapp_demo/dao/redis"
	"webapp_demo/logger"
	"webapp_demo/routes"
	"webapp_demo/settings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

//Go Web开发较为通用的脚手架模版

func main() {
	//1.加载配置文件
	if err := settings.Init(); err != nil {
		fmt.Printf("init setting failed,err:%v\n", err)
		return
	}
	//2.初始化日志
	if err := logger.Init(); err != nil {
		fmt.Printf("init logger failed,err:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success!!!!")
	//3.初始化mysql链接
	if err := mysql.Init(); err != nil {
		fmt.Printf("init mysql failed,err:%v\n", err)
		return
	}
	defer mysql.Close()
	//4.初始化redis链接
	if err := redis.Init(); err != nil {
		fmt.Printf("init redis failed,err:%v\n", err)
		return
	}
	defer redis.Close()
	//5.注册路由
	r := routes.SetUp()
	//6.启动服务（优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler: r,
	}
	go func() {
		//开启一个gorountine服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	//等待中断信号来优雅的关闭服务器，为关闭服务器操作设置一个5s的超时时间
	quit := make(chan os.Signal, 1) //创建一个接收信号通道
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("shutdown server ....")
	ctx, cancal := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancal()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown", zap.Error(err))

	}
	zap.L().Info("server exiting")
}
