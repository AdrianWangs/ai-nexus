package dal

import (
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/dal/mysql"
	"github.com/AdrianWangs/ai-nexus/go-service/nexus/biz/dal/redis"
)

func Init() {
	redis.Init()
	mysql.Init()
}
