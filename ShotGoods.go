package main

import (
	"bytes"
	"encoding/json"
	"flag"
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
	XRPCVerifyKey   string `json:"x-rpc-verify_key"`
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

// determine input time type (time string or timestamp)
func isOnlyDigital(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// timeExceed format: Year-Month-Day Hour:Minute:Second
// hgd: support Year-Month-Day format for pre-running
// func getTime(timeExceed string) time.Time {
func getTime(timeExceed string) time.Time {
	timeArray := strings.Split(timeExceed, " ")

	formatTime := strings.Split(timeArray[0], "-")
	formatTime = append(formatTime, strings.Split(timeArray[1], ":")...)

	exYear, _ := strconv.Atoi(formatTime[0])
	exMonth, _ := strconv.Atoi(formatTime[1])
	exDay, _ := strconv.Atoi(formatTime[2])
	exHour, _ := strconv.Atoi(formatTime[3])
	exMin, _ := strconv.Atoi(formatTime[4])
	exSec, _ := strconv.Atoi(formatTime[5])
	loca, _ := time.LoadLocation("Asia/Shanghai")
	return time.Date(exYear, time.Month(exMonth), exDay, exHour, exMin, exSec, 0, loca)
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
	// GHL: RTT delay should be dived into 2, because we can only get the double trip delay.
	// so we assume each delay of two round trip is the same.
	return (sum / int64(succ)) / 2
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
func ReadConfig(file string) {
	// 读取配置
	f, err := os.ReadFile(file)
	if err != nil {
		log.Fatal("配置文件读取错误: ", err)
	}
	err = json.Unmarshal(f, &config)
	if err != nil {
		log.Fatal("JSON格式错误: ", err)
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
	var conf string
	var targetTime string
	var goodID string
	var goodType string
	var uid string
	var region string
	var game_biz string

	flag.StringVar(&conf, "config", "", "The config file")
	flag.StringVar(&targetTime, "target", "", "The target time (recommend \"Year-Month-Day Hour:Minute:Second\" as input)")
	flag.StringVar(&goodID, "id", "", "The good ID")
	flag.StringVar(&goodType, "type", "", "The good type(virtual / real)")
	flag.StringVar(&uid, "uid", "", "Yuanshen uid")
	flag.StringVar(&region, "region", "", "ServerRegion(Genshin:cn_gf01 / Honkai3:pc01)")
	flag.StringVar(&game_biz, "game_biz", "", "(Genshin:hk4e_cn / Honkai3:bh3_cn)")

	flag.Parse()
	if targetTime == "" {
		log.Fatal("no target time input")
	}
	if goodID == "" {
		log.Fatal("no good id input")
	}
	ReadConfig(conf)
	var good *Goods
	switch goodType {
	case "virtual":
		if uid == "" {
			log.Fatal("no uid input")
		}
		if region == "" {
			log.Fatal("no region input, (Genshin:cn_gf01 / Honkai3:pc01)")
		}
		if game_biz == "" {
			log.Fatal("no game_biz input, (Genshin:hk4e_cn / Honkai3:bh3_cn)")
		}
		good = NewVirtualGood(goodID, 1, uid, region, game_biz)
	case "real":
		if config.AddressId == 0 {
			log.Fatal("no address input")
		}
		good = NewRealGood(goodID, 1, config.AddressId)
	default:
		log.Fatal("invalid good type! please input virtual / real")
	}

	// Determine if the input time consists of only digits
	if isOnlyDigital(targetTime) {
		good.Worker(parseUnix(targetTime))
	} else {
		good.Worker(getTime(targetTime))
	}
}