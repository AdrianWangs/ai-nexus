package service

import (
	"context"
	user_microservice "github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
	"testing"
)

func TestThirdPartyLogin_Run(t *testing.T) {
	ctx := context.Background()
	s := NewThirdPartyLoginService(ctx)
	// init req and assert value

	request := &user_microservice.ThirdPartyLoginRequest{}
	resp, err := s.Run(request)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
