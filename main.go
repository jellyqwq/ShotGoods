package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

func main() {
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
		if count == 1 {
			break
		}
	}
	for time.Now().UnixNano() < 1678878002000000000 {
		<-ch
	}
	endTime := time.Since(start)
	fmt.Printf("cost: %v", endTime)
}

func shot(GoodsForm Goods, ch chan bool) {
	fmt.Printf("shot time: %v\n", time.Now().UnixMicro())
	headers := map[string]string{
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "en-US,en-GB;q=0.9,en;q=0.8",
		"Accept-Encoding":    "gzip, deflate, br",
		"Content-Type":       "application/json;charset=utf-8",
		// x-rpc参数填自己的
		"x-rpc-device_model": "",
		"x-rpc-device_fp":    "",
		"x-rpc-client_type":  "",
		"x-rpc-device_id":    "",
		"x-rpc-channel":      "",
		"x-rpc-app_version":  "",
		"x-rpc-device_name":  "",
		"x-rpc-sys_version":  "",
		"Origin":             "https://webstatic.miyoushe.com",
		"Referer":            "https://webstatic.miyoushe.com/",
		// ua和cookie也填自己的 必要cookie是account_id和cookie_token
		"User-Agent":         "",
		"Cookie":             "",
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	jsonFrom, err := json.Marshal(GoodsForm)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
	}
	formBytesReader := bytes.NewReader(jsonFrom)
	req, err := http.NewRequest("POST", "https://api-takumi.miyoushe.com/mall/v1/web/goods/exchange", formBytesReader)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Got error %s", err.Error())
	}
	fmt.Println(string(body))
	ch <- true
}
