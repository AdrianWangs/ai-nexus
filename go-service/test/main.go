// @Author Adrian.Wang 2024/8/2 下午3:59:00
package main

import (
	"context"
	"fmt"
	"github.com/AdrianWangs/ai-nexus/go-common/nacos"
	"github.com/AdrianWangs/ai-nexus/go-service/user/kitex_gen/user_microservice"
	"github.com/AdrianWangs/ai-nexus/go-service/user/kitex_gen/user_microservice/userservice"
	"github.com/cloudwego/kitex/client"
	"github.com/kitex-contrib/registry-nacos/resolver"
	"time"
)

func main() {

	configClient, err := nacos.GetNacosConfigClient()

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	r := resolver.NewNacosResolver(configClient)

	cli := userservice.MustNewClient(
		"user-service",
		client.WithResolver(r),
		client.WithRPCTimeout(3*time.Second),
	)

	get_user, err := cli.GetUser(context.Background(), &user_microservice.GetUserRequest{
		UserId: 1,
	})
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println("结果：", get_user)

}
