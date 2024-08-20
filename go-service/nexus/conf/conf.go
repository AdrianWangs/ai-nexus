package conf

import (
	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/kr/pretty"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"os"
	"sync"

	common_config "github.com/AdrianWangs/ai-nexus/go-common/conf"
	"github.com/cloudwego/kitex/pkg/klog"
	"gopkg.in/yaml.v2"
)

var (
	conf *Config
	once sync.Once
)

type Config struct {
	Env      string
	Kitex    Kitex    `yaml:"kitex"`
	MySQL    MySQL    `yaml:"mysql"`
	Redis    Redis    `yaml:"redis"`
	Registry Registry `yaml:"registry"`
	Nacos    Nacos    `yaml:"nacos"`
}

type MySQL struct {
	DSN string `yaml:"dsn"`
}

type Redis struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type Kitex struct {
	Service         string `yaml:"service"`
	Address         string `yaml:"address"`
	EnablePprof     bool   `yaml:"enable_pprof"`
	EnableGzip      bool   `yaml:"enable_gzip"`
	EnableAccessLog bool   `yaml:"enable_access_log"`
	LogLevel        string `yaml:"log_level"`
	LogFileName     string `yaml:"log_file_name"`
	LogMaxSize      int    `yaml:"log_max_size"`
	LogMaxBackups   int    `yaml:"log_max_backups"`
	LogMaxAge       int    `yaml:"log_max_age"`
}

type Registry struct {
	RegistryAddress []string `yaml:"registry_address"`
	Username        string   `yaml:"username"`
	Password        string   `yaml:"password"`
}

type Nacos struct {
	Address             string `yaml:"address"`
	Port                uint64 `yaml:"port"`
	Namespace           string `yaml:"namespace"`
	Group               string `yaml:"group"`
	Username            string `yaml:"username"`
	Password            string `yaml:"password"`
	LogDir              string `yaml:"log_dir"`
	CacheDir            string `yaml:"cache_dir"`
	LogLevel            string `yaml:"log_level"`
	TimeoutMs           uint64 `yaml:"timeout_ms"`
	NotLoadCacheAtStart bool   `yaml:"not_load_cache_at_start"`
}

// GetConf gets configuration instance
func GetConf() *Config {
	once.Do(initConf)
	return conf
}

func initConf() {

	// 获取当前环境配置
	env := GetEnv()
	klog.Infof("当前环境: %s", env)

	conf = new(Config)

	// 从公共配置中加载 Nacos 配置
	nacos_config := common_config.GetConf().Nacos
	client, err := nacos.NewClient(nacos.Options{
		Address:     nacos_config.Address,
		Port:        nacos_config.Port,
		NamespaceID: nacos_config.Namespace,
		Group:       nacos_config.Group,
	})

	if err != nil {
		panic(err)
	}
	client.RegisterConfigCallback(vo.ConfigParam{
		DataId:   "nexus-config.yaml",
		Group:    env,
		Type:     "yaml",
		OnChange: nil,
	}, func(s string, parser nacos.ConfigParser) {
		err = yaml.Unmarshal([]byte(s), conf)
		if err != nil {
			klog.Error("转换配置失败 - %v", err)
			panic(err)
		}
		klog.Info("重启配置")

		// 打印配置
		klog.Infof("配置:")

		pretty.Printf("%# v\n", conf)

		klog.Info("配置加载成功")
	}, 100)
}

func GetEnv() string {
	e := os.Getenv("GO_ENV")
	if len(e) == 0 {
		return "test"
	}
	return e
}

func LogLevel() klog.Level {
	level := GetConf().Kitex.LogLevel
	switch level {
	case "trace":
		return klog.LevelTrace
	case "debug":
		return klog.LevelDebug
	case "info":
		return klog.LevelInfo
	case "notice":
		return klog.LevelNotice
	case "warn":
		return klog.LevelWarn
	case "error_code":
		return klog.LevelError
	case "fatal":
		return klog.LevelFatal
	default:
		return klog.LevelInfo
	}
}
