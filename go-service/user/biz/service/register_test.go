package service

import (
	"context"
	"github.com/AdrianWangs/nexus/go-service/user/biz/dal/mysql"
	"github.com/AdrianWangs/nexus/go-service/user/kitex_gen/user_microservice"
	"testing"
)

func TestRegister_Run(t *testing.T) {

	ctx := context.Background()
	s := NewRegisterService(ctx)
	// init req and assert value

	mysql.Init()

	request := &user_microservice.RegisterRequest{
		Username:    "test",
		Password:    "test",
		Birthday:    "2020-7-10",
		Gender:      "test",
		Email:       "test@test",
		PhoneNumber: "1234567890",
	}
	resp, err := s.Run(request)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test

}
