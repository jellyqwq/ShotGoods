package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	RETRIES  = 5
	TESTS    = 3
	TRY_SEND = 10
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

var (
	headers = map[string]string{
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "en-US,en-GB;q=0.9,en;q=0.8",
		"Accept-Encoding": "gzip, deflate, br",
		"Content-Type":    "application/json;charset=utf-8",
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
		"User-Agent": "",
		"Cookie":     "",
	}
)

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
	h, m, s := t.Clock()
	exHour, _ := strconv.Atoi(formatTime[0])
	exMin, _ := strconv.Atoi(formatTime[1])
	exSec, _ := strconv.Atoi(formatTime[2])
	return t.Add((time.Duration(exHour - h)) * time.Hour).Add(time.Duration(exMin-m) * time.Minute).Add(time.Duration(exSec-s) * time.Second)
}

func NewGood(app_id, exchange_num uint8, address_id uint32, point_sn, goods_id string) *Goods {
	return &Goods{
		Address_id:   address_id,
		App_id:       app_id,
		Exchange_num: exchange_num,
		Goods_id:     goods_id,
		Point_sn:     point_sn,
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
	var t1 time.Time
	for i := 0; i < TESTS; i++ {
		for j := 0; j < RETRIES && !ok; j++ {
			t1 = time.Now()
			ok, _ = g.doRequest(false)
		}
		sum += time.Since(t1).Milliseconds()
	}
	return sum / TESTS
}

func (g *Goods) GrabIt() {
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
	wg.Wait()
}

func (g *Goods) Worker(timeExceed time.Time) {
	commCheck := time.NewTicker(time.Second)
	var highFrequent *time.Ticker
	hasProbe := false
	var latency int64
	var realExceed time.Time
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
				g.GrabIt()
			}
			if now.After(realExceed) {
				g.GrabIt()
				highFrequent.Stop()
				return
			}
		}
	}
}
func main() {
	good := NewGood(1, 1, 0, "myb", "2023022311902")

	good.Worker(getTime("19:00:00"))
}