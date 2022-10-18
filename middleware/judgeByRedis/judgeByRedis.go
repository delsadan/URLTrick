package judgeByRedis

import (
	"URLTrick/config"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

var rdb *redis.Client

func initRedis() (err error) {
	MyConfig, err := config.ReadConfig()
	if err != nil {
		return err
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     MyConfig.RedisConfig.Addr,
		Password: MyConfig.RedisConfig.Password,
		DB:       MyConfig.RedisConfig.Db,
	})
	_, err = rdb.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func JudgeKeyByRedis(key string, limitSeconds int) (keyNum int64) {
	keyNum = rdb.Incr(key).Val()

	if limitSeconds == 0 {
		return
	}
	if keyNum == 1 {
		rdb.Expire(key, time.Second*time.Duration(limitSeconds))
	}
	return keyNum
}

func DelKeyByRedis(key string) {
	rdb.Del(key)
}

func main() {
	//err := initRedis()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//h, _ := rdb.Set("hello", "world", time.Second*60).Result()

	n := JudgeKeyByRedis("b", 60)
	fmt.Println(n)

}

func init() {
	err := initRedis()
	if err != nil {
		panic(err)
	}
	DelKeyByRedis("t")
}
