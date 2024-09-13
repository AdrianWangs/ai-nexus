package main

import (
	"github.com/AdrianWangs/ai-nexus/go-common/nacos"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/dal"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/conf"
	nexus_microservice "github.com/AdrianWangs/ai-nexus/go-service/nexus/kitex_gen/nexus_microservice/nexusservice"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/joho/godotenv"
	kitexlogrus "github.com/kitex-contrib/obs-opentelemetry/logging/logrus"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {

	// 读取环境变量
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("环境变量文件加载失败", err)
	}

	opts := kitexInit()

	svr := nexus_microservice.NewServer(new(NexusServiceImpl), opts...)

	err = svr.Run()

	if err != nil {
		log.Println(err.Error())
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

	logger.Logger().SetReportCaller(true)
	logger.Logger().SetFormatter(&MyFormatter{})

	klog.SetLogger(logger)
	klog.SetLevel(klog.LevelDebug)

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
