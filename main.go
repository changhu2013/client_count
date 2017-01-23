package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

func getRedisConnect(redisURL string) redis.Conn {
	var c, e = redis.DialURL(redisURL)

	if e != nil {
		log.Fatal(e)
		return nil
	}

	return c
}

func doGetClientCount(c redis.Conn, key string) int64 {
	var v, e = redis.String(c.Do("GET", key))

	if e != nil {
		return 0
	}

	cc, ee := strconv.ParseInt(v, 10, 0)

	if ee != nil {
		return 0
	}

	return cc
}

func getClientCount(conn redis.Conn, vvv bool) int64 {
	var ks, e = redis.Strings(conn.Do("KEYS", "websocket_clients_count_*"))

	if e != nil {
		log.Fatal(e)
		return 0
	}

	var c int64

	for _, k := range ks {
		var cc = doGetClientCount(conn, k)

		if vvv {
			fmt.Println(k, cc)
		}

		c = c + cc
	}

	return c
}

func doReport(monitorURL string, count int64, step int64) {
	json := `[{"metric":"%s","endpoint":"%s","timestamp":%d,"step":%d,"value":%d,"counterType":"GAUGE","tags":"%s"}]`

	metric := "beeper_mpp_connection_count"
	hostname, _ := os.Hostname()
	timestamp := time.Now().Unix()
	tags := "project=beeper_mpp,module=master,value=client_count"

	content := fmt.Sprintf(json, metric, hostname, timestamp, step, count, tags)

	log.Println(content)

	client := &http.Client{}
	req, err := http.NewRequest("POST", monitorURL, bytes.NewBufferString(content))

	if err != nil {
		log.Fatal(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Println(string(body))
	}
}

func main() {

	var h bool
	var v bool
	var s string
	var r string
	var m string

	flag.BoolVar(&h, "h", false, "帮助信息")
	flag.BoolVar(&v, "v", false, "显示更多细节信息")
	flag.StringVar(&s, "s", "10", "向监控系统实时发送连接数的频率，单位(秒)")
	flag.StringVar(&r, "r", "redis://127.0.0.1:6379", "Redis连接URL")
	flag.StringVar(&m, "m", "http://127.0.0.1:9090", "监控系统服务地址")

	flag.Parse()

	if h {
		flag.PrintDefaults()
		return
	}

	ss, err := time.ParseDuration(s + "s")

	if err != nil {
		log.Fatal(err)
		return
	}

	var conn = getRedisConnect(r)
	defer conn.Close()

	var step, _ = strconv.ParseInt(s, 10, 0)

start:

	doReport(m, getClientCount(conn, v), step)
	time.Sleep(ss)

	goto start
}
