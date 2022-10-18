package main

import (
	"URLTrick/config"
	"URLTrick/middleware/judgeByRedis"
	"URLTrick/middleware/pushWX"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"sync"
	"time"
)

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

var startTickerSignal int = 1
var Keys = make(map[string]int)
var mu sync.RWMutex

func VerityUrlParamHandler(c *gin.Context) {
	if strings.Contains(c.Param(c.MustGet("indexUrl").(string)), c.MustGet("urlParameterKey").(string)) {
		cCp := c.Copy()
		urlParameterKey := cCp.MustGet("urlParameterKey").(string)
		urlParameterTimeLimitSeconds := cCp.MustGet("urlParameterTimeLimitSeconds").(int)
		urlParameterThreshold := cCp.MustGet("urlParameterThreshold").(int64)
		go func() {
			reverse := cCp.MustGet("reverse").(bool)
			if reverse {
				if startTickerSignal == 1 {
					mu.Lock()
					Keys[Md5(urlParameterKey)] = 0
					mu.Unlock()
					go func() {
						t := time.NewTicker(time.Second * time.Duration(urlParameterTimeLimitSeconds))
						<-t.C
						startTickerSignal = 1
						mu.RLock()
						num := Keys[Md5(urlParameterKey)]
						mu.RUnlock()
						if num < int(urlParameterThreshold) { // 小于预设值告警
							go func() {
								message, err := pushWX.NewMessage(fmt.Sprintf("发现阈值小于预设值告警: %s \n当前值: %d \n预设值: %d \n访问端: %s \n", urlParameterKey, num, urlParameterThreshold, cCp.ClientIP()), 1)
								if err != nil {
									return
								}
								message.PushMessageToWX()
							}()
						}
					}()
				}
				mu.Lock()
				Keys[Md5(urlParameterKey)] += 1
				mu.Unlock()
				startTickerSignal = 0
			} else {
				num := judgeByRedis.JudgeKeyByRedis(Md5(urlParameterKey),
					urlParameterTimeLimitSeconds)
				if num >= urlParameterThreshold {
					judgeByRedis.DelKeyByRedis(Md5(urlParameterKey))
					message, err := pushWX.NewMessage(fmt.Sprintf("发现阈值: %s 大于或等于预设值告警 \n当前值:%d \n预设值: %d \n访问端: %s \n", urlParameterKey, num, urlParameterThreshold, cCp.ClientIP()), 1)
					if err != nil {
						return
					}
					message.PushMessageToWX()
				}
			}
		}()

		c.String(200, "ok")
		return
	}

	c.String(200, "ok")
}

func VerityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		myConfig, err := config.ReadConfig()
		if err != nil {
			c.Abort()
			c.String(400, "server internal fail")
			return
		}
		c.Set("indexUrl", myConfig.IndexUrl) // 提供配置文件的indexUrl
		c.Set("reverse", myConfig.Reverse)
		for _, trigger := range myConfig.Triggers {

			if trigger.Position == "header" {
				c.Set("xHeaderKey", trigger.Key)
				c.Set("xHeaderThreshold", trigger.Threshold)
				c.Set("xHeaderLimitSeconds", trigger.TimeLimitSeconds)

			} else if trigger.Position == "urlParameter" {
				c.Set("urlParameterKey", trigger.Key)
				c.Set("urlParameterThreshold", trigger.Threshold)
				c.Set("urlParameterTimeLimitSeconds", trigger.TimeLimitSeconds)

			}
		}

		c.Next()
	}
}

func VerityHeaderHandler(c *gin.Context) {
	cCp := c.Copy()
	xHeaderKey := cCp.MustGet("xHeaderKey").(string)
	newXHeaderKeyArrary := strings.Split(xHeaderKey, "=")
	xHeader, isPresent := c.Request.Header[newXHeaderKeyArrary[0]]
	if !isPresent {
		c.String(200, "fail")
		return
	} else if xHeader[0] == newXHeaderKeyArrary[1] {
		go func() {
			xHeaderThreshold := cCp.MustGet("xHeaderThreshold").(int64)
			xHeaderLimitSeconds := cCp.MustGet("xHeaderLimitSeconds").(int)

			reverse := cCp.MustGet("reverse").(bool)

			if reverse { //反选
				if startTickerSignal == 1 {
					mu.Lock()
					Keys[Md5(xHeaderKey)] = 0
					mu.Unlock()
					go func() {
						t := time.NewTicker(time.Second * time.Duration(xHeaderLimitSeconds))
						<-t.C
						startTickerSignal = 1
						mu.RLock()
						num := Keys[Md5(xHeaderKey)]
						mu.RUnlock()
						if num < int(xHeaderThreshold) { // 小于预设值告警
							go func() {
								message, err := pushWX.NewMessage(fmt.Sprintf("发现阈值小于预设值告警: %s \n当前值:%d \n预设值: %d \n访问端: %s \n", xHeaderKey, num, xHeaderThreshold, cCp.ClientIP()), 1)
								if err != nil {
									return
								}
								message.PushMessageToWX()
							}()
						}
					}()
				}
				mu.Lock()
				Keys[Md5(xHeaderKey)] += 1
				mu.Unlock()
				startTickerSignal = 0

			} else {
				num := judgeByRedis.JudgeKeyByRedis(Md5(newXHeaderKeyArrary[1]), xHeaderLimitSeconds)
				if num >= xHeaderThreshold {
					judgeByRedis.DelKeyByRedis(Md5(newXHeaderKeyArrary[1]))
					message, err := pushWX.NewMessage(fmt.Sprintf("发现阈值: %s 大于或等于预设值告警 \n当前值:%d \n预设值: %d \n访问端: %s \n", newXHeaderKeyArrary, num, xHeaderThreshold, cCp.ClientIP()), 1)
					if err != nil {
						return
					}
					message.PushMessageToWX()

				}
			}
		}()

		c.JSON(200, "ok")
		//c.String(200, "ok")
		return
	}
	c.String(200, "ok")
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	err := r.SetTrustedProxies(nil)
	if err != nil {
		return
	}
	myConfig, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	for _, trigger := range myConfig.Triggers {
		if trigger.Position == "header" {
			if trigger.Key == "" {
				fmt.Println("未启用header判断")
			} else {
				r.GET(fmt.Sprintf("/%s", myConfig.IndexUrl), VerityMiddleware(), VerityHeaderHandler)
			}

		} else if trigger.Position == "urlParameter" {
			if trigger.Key == "" {
				fmt.Println("未启用urlParameter判断")
			} else {
				r.GET(fmt.Sprintf("/%s/*%s", myConfig.IndexUrl, myConfig.IndexUrl), VerityMiddleware(), VerityUrlParamHandler)
			}
		}
	}
	r.GET("/", func(context *gin.Context) {
		context.String(200, context.ClientIP())
	})
	err = r.Run(myConfig.ListenHost)
	if err != nil {
		return
	}
}
