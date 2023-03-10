package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	RETRIES  = 5
	TESTS    = 3
	TRY_SEND = 5
)

var (
	config  Config
	headers map[string]string
)

type Goods struct {
	// GHL: Why use uint8? Confused.
	App_id       uint8  `json:"app_id"`
	Point_sn     string `json:"point_sn"`
	Goods_id     string `json:"goods_id"`
	Exchange_num uint8  `json:"exchange_num"`
	Uid          string `json:"uid,omitempty"`
	Region       string `json:"region,omitempty"`
	Game_biz     string `json:"game_biz,omitempty"`
	// GHL: why uint32???
	Address_id uint32 `json:"address_id"`
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
	AddressId       uint32 `json:"address_id"`
}

// TODO: custom config
func setHeader(h http.Header) {
	for k, v := range headers {
		// use Set instead of Add, Add may not be overrided.
		h.Set(k, v)
	}
}

// timeExceed format: Hour:Minute:Second
func getTime(timeExceed string) time.Time {
	formatTime := strings.Split(timeExceed, ":")
	t := time.Now()
	exHour, _ := strconv.Atoi(formatTime[0])
	exMin, _ := strconv.Atoi(formatTime[1])
	exSec, _ := strconv.Atoi(formatTime[2])
	loca, _ := time.LoadLocation("Asia/Shanghai")
	return time.Date(t.Year(), t.Month(), t.Day(), exHour, exMin, exSec, 0, loca)
}

func parseUnix(unix string) time.Time {
	ut, _ := strconv.ParseInt(unix, 10, 64)
	return time.Unix(ut, 0)
}

func NewRealGood(goods_id string, exchange_num uint8, address_id uint32) *Goods {
	// 兑换实物奖励
	return &Goods{
		App_id:       1,
		Point_sn:     "myb",
		Goods_id:     goods_id,
		Exchange_num: exchange_num,
		Address_id:   address_id,
	}
}

func NewVirtualGood(goods_id string, exchange_num uint8, uid, region, game_biz string) *Goods {
	// 兑换游戏奖励
	return &Goods{
		App_id:       1,
		Point_sn:     "myb",
		Goods_id:     goods_id,
		Exchange_num: exchange_num,
		Uid:          uid,
		Region:       region,
		Game_biz:     game_biz,
		Address_id:   0,
	}
}

func (g *Goods) Byte() []byte {
	jsonFrom, _ := json.Marshal(g)
	return jsonFrom
}

func (g *Goods) doRequest(needRead ...bool) (ok bool, ret []byte) {
	isRequireContent := true
	if len(needRead) > 0 {
		isRequireContent = needRead[0]
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	formBytesReader := bytes.NewReader(g.Byte())
	req, err := http.NewRequest("POST", "https://api-takumi.miyoushe.com/mall/v1/web/goods/exchange", formBytesReader)
	if err != nil {
		// GHL: Replace fmt with log to format the log contents
		// Don't continue to execute when the error occurred!!!
		log.Printf("Got error %s", err.Error())
		return
	}
	setHeader(req.Header)
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Got error %s", err.Error())
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return
	}
	ok = true
	if isRequireContent {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Printf("Got error %s", err.Error())
			return
		}
		ret = body
	}
	return
}

func (g *Goods) TestLatency() int64 {
	var ok bool
	var sum int64
	var delay int64
	var prev int64
	var succ int
	var t1 time.Time
	for i := 0; i < TESTS; i++ {
		ok = false
		for j := 0; j < RETRIES && !ok; j++ {
			t1 = time.Now()
			ok, _ = g.doRequest(false)
		}
		delay = time.Since(t1).Milliseconds()
		// diff is too large, skip it.
		if prev > 0 && delay-prev >= 500 {
			continue
		}
		prev = delay
		if !ok {
			continue
		}
		sum += delay
		succ += 1
	}
	return sum / int64(succ)
}

func (g *Goods) GrabIt() *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(TRY_SEND)
	for i := 0; i < TRY_SEND; i++ {
		go func() {
			defer wg.Done()
			ok, ret := g.doRequest()
			if ok {
				log.Println(string(ret))
			} else {
				log.Println("Requests Fail")
			}
		}()
	}
	return &wg
}

func (g *Goods) Worker(timeExceed time.Time) {
	commCheck := time.NewTicker(time.Second)
	var highFrequent *time.Ticker
	hasProbe := false
	hasSent := false
	var latency int64
	var realExceed time.Time
	waitGroup := []*sync.WaitGroup{}
	defer func() {
		for _, wg := range waitGroup {
			wg.Wait()
		}
	}()
	for {
		if !hasProbe {
			now := <-commCheck.C
			if timeExceed.Sub(now).Seconds() < 5 {
				log.Println("Start to test the latency")
				latency = g.TestLatency()
				if latency == 0 {
					log.Println("Probe fail")
					continue
				}
				log.Printf("Test finished: delay %d", latency)
				hasProbe = true
				commCheck.Stop()
				highFrequent = time.NewTicker(10 * time.Millisecond)
				packet_delay := time.Duration(latency) * time.Millisecond
				realExceed = timeExceed.Add(-packet_delay)
			}
		} else {
			now := <-highFrequent.C
			if now.Equal(realExceed) {
				waitGroup = append(waitGroup, g.GrabIt())
			}
			if now.After(realExceed) && !hasSent {
				waitGroup = append(waitGroup, g.GrabIt())
				hasSent = true
			}
			if now.Equal(timeExceed) || now.After(timeExceed) {
				waitGroup = append(waitGroup, g.GrabIt())
				highFrequent.Stop()
				return
			}
		}
	}
}
func ReadConfig() {
	// 读取配置
	f, err := os.ReadFile("./config.json")
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(f, &config)
	if err != nil {
		log.Println(err)
		return
	}
	headers = map[string]string{
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "en-US,en-GB;q=0.9,en;q=0.8",
		"Accept-Encoding": "gzip, deflate, br",
		"Content-Type":    "application/json;charset=utf-8",
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
		"User-Agent": config.UserAgent,
		"Cookie":     config.Cookie,
	}
}
func main() {
	ReadConfig()
	// https://github.com/jellyqwq/ShotGoods/blob/main/goods.csv
	// 实物兑换
	// good := NewRealGood("2023022311902", 1, config.AddressId)
	// 游戏内兑换 (原神为例)
	good := NewVirtualGood("2023022412691", 1, "Yuanshen uid", "cn_gf01", "hk4e_cn")
	//good.Worker(getTime("19:00:00"))
	good.Worker(parseUnix("1678878000"))
}
