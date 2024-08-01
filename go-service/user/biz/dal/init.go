package dal

import (
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/mysql"
	"github.com/AdrianWangs/ai-nexus/go-service/user/biz/dal/redis"
)

func Init() {
	redis.Init()
	mysql.Init()
}
