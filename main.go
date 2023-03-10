package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"os"
)

type Goods struct {
	App_id       uint8  `json:"app_id"`
	Point_sn     string `json:"point_sn"`
	Goods_id     string `json:"goods_id"`
	Exchange_num uint8  `json:"exchange_num"`
	Uid          string `json:"uid,omitempty"`
	Region       string `json:"region,omitempty"`
	Game_biz     string `json:"game_biz,omitempty"`
	Address_id   uint32 `json:"address_id"`
}

type Config struct {
	XRPCDeviceModel string `json:"x-rpc-device_model"`
	XRPCDeviceFp    string `json:"x-rpc-device_fp"`
	XRPCClientType  string `json:"x-rpc-client_type"`
	XRPCDeviceID    string `json:"x-rpc-device_id"`
	XRPCChannel     string `json:"x-rpc-channel"`
	XRPCAppVersion  string `json:"x-rpc-app_version"`
	XRPCDeviceName  string `json:"x-rpc-device_name"`
	XRPCSysVersion  string `json:"x-rpc-sys_version"`
	UserAgent       string `json:"User-Agent"`
	Cookie          string `json:"Cookie"`
}

var config Config
var headers map[string]string

func main() {
	// 读取配置
	f, err := os.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(f, &config)
	if err != nil {
		fmt.Println(err)
		return
	}
	headers = map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "en-US,en-GB;q=0.9,en;q=0.8",
		"Accept-Encoding":    "gzip, deflate, br",
		"Content-Type":       "application/json;charset=utf-8",
		// x-rpc参数填自己的
		"x-rpc-device_model": config.XRPCDeviceModel,
		"x-rpc-device_fp":    config.XRPCDeviceFp,
		"x-rpc-client_type":  config.XRPCClientType,
		"x-rpc-device_id":    config.XRPCDeviceID,
		"x-rpc-channel":      config.XRPCChannel,
		"x-rpc-app_version":  config.XRPCAppVersion,
		"x-rpc-device_name":  config.XRPCDeviceName,
		"x-rpc-sys_version":  config.XRPCSysVersion,
		"Origin":             "https://webstatic.miyoushe.com",
		"Referer":            "https://webstatic.miyoushe.com/",
		// ua和cookie也填自己的 必要cookie是account_id和cookie_token
		"User-Agent":         config.UserAgent,
		"Cookie":             config.Cookie,
	}

	ch := make(chan bool)
	start := time.Now()
	count := 0
	for time.Now().UnixNano() < 1678878002000000000 {
		// 需要更换的地方, Goods结构, Address_id在兑换虚拟奖品时为0, 实物兑换需要填上对应id
		// 兑换虚拟物品时要添加三个游戏参数, Uid, Region, Game_biz
		go shot(Goods{
			App_id: 1,
			Point_sn: "myb",
			Exchange_num: 1,
			Goods_id: "2023022311902",
			Address_id: 0,
		}, ch)
		time.Sleep(time.Microsecond*1000)
		count += 1
		if count == 2 {
			break
		}
	}
	for count > 1 {
		<-ch
	}
	endTime := time.Since(start)
	fmt.Printf("cost: %v", endTime)
}

func shot(GoodsForm Goods, ch chan bool) {
	fmt.Printf("shot time: %v\n", time.Now().UnixMicro())
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	jsonFrom, err := json.Marshal(GoodsForm)
	if err != nil {
		fmt.Printf("Got error %s\n", err.Error())
		return
	}
	formBytesReader := bytes.NewReader(jsonFrom)
	req, err := http.NewRequest("POST", "https://api-takumi.miyoushe.com/mall/v1/web/goods/exchange", formBytesReader)
	if err != nil {
		fmt.Printf("NewRequest error %s\n", err.Error())
		return
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("client Do error %s\n", err.Error())
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("io ReadAll %s\n", err.Error())
		return
	}
	fmt.Println(string(body))
	ch <- true
}
