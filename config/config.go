package config

import (
	"fmt"
	"github.com/goccy/go-json"
	"io/ioutil"
	"sync"
)

type Trigger struct {
	Key              string `json:"key"`
	Position         string `json:"position"`
	Threshold        int64  `json:"threshold"`
	TimeLimitSeconds int    `json:"timeLimitSeconds"`
}

type Config struct {
	Triggers    []Trigger   `json:"triggers"`
	IndexUrl    string      `json:"indexUrl"`
	RedisConfig redisConfig `json:"redis"`
	WxHelper    wxHelper    `json:"wxHelper"`
	Reverse     bool        `json:"reverse"`
	ListenHost  string      `json:"listenHost"`
}

type redisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

type wxHelper struct {
	AppToken string `json:"appToken"`
	TopicIds []int  `json:"topicIds"`
}

var fileLocker sync.Mutex

func ReadConfig() (Config, error) {
	filename := "./config/config.json"
	var config Config
	fileLocker.Lock()
	data, err := ioutil.ReadFile(filename)
	fileLocker.Unlock()
	if err != nil {
		fmt.Printf("%s 打开失败", filename)
		return config, err
	}
	configJson := []byte(data)
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		fmt.Println("Unmarshal失败")
		return config, err
	}
	return config, nil
}

func main() {
	myConfig, err := ReadConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println(myConfig.IndexUrl, myConfig.Triggers[1].TimeLimitSeconds)
}
