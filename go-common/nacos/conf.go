// @Author Adrian.Wang 2024/7/24 下午1:43:00
package nacos

import (
	"github.com/AdrianWangs/ai-nexus/go-common/conf"
	"github.com/cloudwego/kitex/pkg/discovery"
	register "github.com/cloudwego/kitex/pkg/registry"
	"github.com/kitex-contrib/registry-nacos/registry"
	"github.com/kitex-contrib/registry-nacos/resolver"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"log"
)

func GetNacosConfigClient() (iClient naming_client.INamingClient, err error) {

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
		log.Fatalf("create nacos client error_code: %v", err)
		return nil, err
	}

	return client, nil

}

func GetNacosRegistry() register.Registry {

	// 获取nacos配置
	client, err := GetNacosConfigClient()

	if err != nil {
		log.Fatalf("get nacos client error_code: %v", err)
		return nil
	}

	return registry.NewNacosRegistry(client)

}

// GetNacosResolver 获取nacos resolver
func GetNacosResolver() discovery.Resolver {

	// 获取nacos配置
	client, err := GetNacosConfigClient()

	if err != nil {
		log.Fatalf("get nacos client error_code: %v", err)
		return nil
	}

	return resolver.NewNacosResolver(client)

}
