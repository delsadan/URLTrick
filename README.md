# URLTrick(URL小陷阱)

## 1、 简介
URLTrick使用场景，运维监控，侧信道URL访问监控。
- 支持指定时间内，触发设定阈值指定次数（大于或等于），即微信推送消息告警
- 支持指定时间内，未到设定阈值（小于），即微信推送消息告警


代码逻辑：利用gin框架起一个indexURL，然后识别访问的该路径的路径参数及header头，有符合规则（key）的md5哈希入库（Redis），通过Redis的INCR 和 EXPIRE 实现有效期的阈值监控。
同时没写redis初始化清除键及值，所以会复用redis之前键，若之前有设timeLimitSeconds为0，然后有设有值，后面的不会生效。需去redis里清空该键值。
同时借助WXpusher实现触发阈值，微信推送消息告警的目的。

整体逻辑非常简单，我主要用来看我的小爬虫是否正常运行，还有做一些特殊的隐藏告警

第一次用GO写程序，本来可以不用redis，为了试试redis库，所以故意写了复杂点，可能有bug，凑合用用了。

## 2、 配置解读

```bash
# 此处为介绍，配置文件为JSON格式，不支持注释，请务必删除注释
{
  # 识别字段配置, 暂未支持多组key识别
  "triggers": [ 
    {
      "key": "X-Custom-Num=4", # 留空不启用, 首字母大写, 支持热加载
      "position": "header", # 固定值1, 不支持热加载
      "threshold": 3, # 告警阈值，达到次数及告警, 支持热加载（需等上一个key过期）
      "timeLimitSeconds": 60 # 告警阈值有限期，单位秒，永久设为0, 支持热加载
    },
    {
      "key": "abc", # 识别的url参数，此处为包含识别，如indexUrl/abc、indexUrl/123abcdef，均可识别
      "position": "urlParameter",# 固定值2, 不支持热加载
      "threshold": 3, # 告警阈值，达到次数及告警
      "timeLimitSeconds": 60 # 告警阈值有限期，单位秒，永久设为0
    }
  ],
  "indexUrl": "heartbeat", # 触发url, 暂不支持多级, 勿加url路径分隔符, 不支持热加载 
  # redis配置
  "redis" : {
    "addr": "localhost:6379",
    "password": "",
    "db": 0
  },
  # wxpusher配置, 详见docs: https://wxpusher.dingliqc.com/docs/#/
  "wxhelper": {
    "appToken": "AT_XXXXX", # appToken, 使用前务必修改
    "topicIds": [10000]  # 主题Topic, 使用前务必修改
  },
  # 全局告警反转设置, 默认false，为达到阈值触发告警，反之true，则告警阈值有效期内没有达到告警次数告警，两者的实现方式并不一样，后者的告警需等阈值时间到以后，发送告警消息
  "reverse" : false,  # 支持热加载
  "listenHost" : "0.0.0.0:18088" # 服务端监听端口
}
```

## 3、使用方式
### 3.1、 Centos 7.6 自行编译
```bash 
# 准备Go环境 Go Version > 1.17
sudo yum install -y epel-release
sudo yum install -y golang
# 准备Redis环境
yum install redis # 此安装为3.2.X版本, 别bind 0.0.0.0而不设密码啊
systemctl enable redis
systemctl start redis
# 部署源代码（略）
# 配置config.json, 见上一节
go build -o server server.go
chmod +x server
./server
# 注意: ./config/config.json 必须在server二进制程序目录下


```
### 3.2 使用release
```bash 
# 准备redis环境，见上一节
# 下载release略
# 配置./config/config.json
chmod +x server
./server
```

### 3.3 长期运行(centos service)
```bash
# 将以下文件保存至/usr/lib/systemd/system/urlTrick.service
vi /usr/lib/systemd/system/urlTrick.service
# 复制以下内容
[Unit]
Description=urlTrick
After=network.target

[Service]
Type=simple
# 修改二进制程序所在目录
WorkingDirectory= /root/URLTrick # 修改
ExecStart=/root/URLTrick/server >/dev/null 2>/var/log/URLTrick.log & # 修改, 后面的代表错误日志写入指定log
KillMode=mixed
Restart=on-failure
RestartSec=60s

[Install]
WantedBy=multi-user.target

systemctl start urlTrick
systemctl enable urlTrick  # 开启自启动
```

## 4、使用示例

根据以上配置文件, 则60s内访问3次: http://0.0.0.0:18088/heartbeat/abc, 即可触发微信告警, 或者携带X-Custom-Num=4的headers头访问3次: http://0.0.0.0:18088/heartbeat, 也会触发告警