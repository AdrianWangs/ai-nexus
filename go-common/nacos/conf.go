// @Author Adrian.Wang 2024/7/24 下午1:43:00
package nacos

import (
	"github.com/AdrianWangs/nexus/go-common/conf"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func GetNacosConfig() (iClient naming_client.INamingClient, err error) {

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

	return clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)

}
