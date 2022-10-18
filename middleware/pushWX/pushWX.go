package pushWX

import (
	"URLTrick/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Message struct {
	AppToken    string `json:"appToken"`
	Content     string `json:"content"`
	ContentType int    `json:"contentType"`
	Summary     string `json:"summary"`
	TopicIds    []int  `json:"topicIds"`
	Url         string `json:"url"`
}

func NewMessage(content string, contentType int) (*Message, error) {
	Config, err := config.ReadConfig()
	if err != nil {
		return nil, err
	}
	message := new(Message)
	message.AppToken = Config.WxHelper.AppToken // Token
	message.Content = content                   // 内容纯文本或者是不带HTML的Body标签
	message.ContentType = contentType           // 1为纯文本，2为html
	message.Summary = ""                        // 消息摘要
	message.Url = ""                            // 跳转链接
	message.TopicIds = Config.WxHelper.TopicIds // 推送主题ID
	return message, nil
}

func (message *Message) PushMessageToWX() {
	serverUrl := "https://wxpusher.zjiecode.com/api/send/message"
	messageJson, err := json.Marshal(message)
	if err != nil {
		panic(fmt.Sprintf("消息序列化为JSON时出现错误: %s", err))
	}
	params := bytes.NewReader(messageJson)
	resp, err := http.Post(serverUrl, "application/json", params)
	if err != nil {
		panic(fmt.Sprintf("发送微信消息时出现错误: %s", err))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("解析resp.Body时出现错误: %s", err))
	}
	fmt.Println(string(body))
}

func main() {
	message, err := NewMessage("hello world", 1)
	if err != nil {
		panic(fmt.Sprintf("生成消息出现错误: %s", err))
	}
	if err != nil {
		panic(fmt.Sprintf("生成消息出现错误: %s", err))
	}
	message.PushMessageToWX()
}
