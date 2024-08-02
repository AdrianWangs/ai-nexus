package main

import (
	"github.com/AdrianWangs/ai-nexus/go-common/nacos"
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/AdrianWangs/ai-nexus/go-service/user/conf"
	"github.com/AdrianWangs/ai-nexus/go-service/user/kitex_gen/user_microservice/userservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	kitexlogrus "github.com/kitex-contrib/obs-opentelemetry/logging/logrus"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// @title 用户微服务
// @version 1.0
// @description 用户微服务
// @tag.name 用户服务
func main() {

	// 读取环境变量
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("环境变量文件加载失败", err)
	}

	opts := kitexInit()

	svr := userservice.NewServer(new(UserServiceImpl), opts...)

	err = svr.Run()
	if err != nil {
		klog.Error(err.Error())
	}
}

func kitexInit() (opts []server.Option) {

	opts = append(opts, server.
		WithTransHandlerFactory(&mixTransHandlerFactory{nil}))

	// 初始化数据库
	dal.Init()

	// address
	addr, err := net.ResolveTCPAddr("tcp", conf.GetConf().Kitex.Address)
	if err != nil {
		panic(err)
	}
	opts = append(opts, server.WithServiceAddr(addr))

	// service info
	opts = append(opts, server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: conf.GetConf().Kitex.Service,
	}))

	// nacos 注册中心
	r := nacos.GetNacosRegistry()
	opts = append(opts, server.WithRegistry(r))

	// thrift meta handler
	opts = append(opts, server.WithMetaHandler(transmeta.ServerTTHeaderHandler))

	// klog
	logger := kitexlogrus.NewLogger()
	klog.SetLogger(logger)
	klog.SetLevel(conf.LogLevel())
	asyncWriter := &zapcore.BufferedWriteSyncer{
		WS: zapcore.AddSync(&lumberjack.Logger{
			Filename:   conf.GetConf().Kitex.LogFileName,
			MaxSize:    conf.GetConf().Kitex.LogMaxSize,
			MaxBackups: conf.GetConf().Kitex.LogMaxBackups,
			MaxAge:     conf.GetConf().Kitex.LogMaxAge,
		}),
		FlushInterval: time.Minute,
	}

	// 创建一个 MultiWriter，同时写入文件和控制台
	multiWriter := io.MultiWriter(asyncWriter, os.Stdout)

	klog.SetOutput(multiWriter)
	server.RegisterShutdownHook(func() {
		asyncWriter.Sync()
	})
	return
}
