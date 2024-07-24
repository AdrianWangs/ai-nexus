// @Author Adrian.Wang 2024/7/24 下午1:43:00
package nacos

import (
	"github.com/AdrianWangs/nexus/go-common/conf"
	register "github.com/cloudwego/kitex/pkg/registry"
	"github.com/kitex-contrib/registry-nacos/registry"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
)

func GetNacosConfig() (iClient naming_client.INamingClient) {

	nacos := conf.GetConf().Nacos

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(nacos.Address, nacos.Port),
	}

	cc := constant.ClientConfig{
		NamespaceId:         nacos.Namespace,
		TimeoutMs:           nacos.TimeoutMs,
		NotLoadCacheAtStart: nacos.NotLoadCacheAtStart,
		LogDir:              nacos.LogDir,
		CacheDir:            nacos.CacheDir,
		LogLevel:            nacos.LogLevel,
		Username:            nacos.Username,
		Password:            nacos.Password,
	}

	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)

	if err != nil {
		log.Fatalf("create nacos client error: %v", err)
		return nil
	}

	return client

}

func GetNacosRegistry() register.Registry {

	// 获取nacos配置
	client := GetNacosConfig()

	return registry.NewNacosRegistry(client)

}
