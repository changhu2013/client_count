package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

func getRedisConnect(redisURL string) redis.Conn {
	conn, err := redis.DialURL(redisURL)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return conn
}

func doGetClientCount(c redis.Conn, key string) int64 {
	value, err := redis.String(c.Do("GET", key))
	if err != nil {
		return 0
	}
	cc, ee := strconv.ParseInt(value, 10, 0)
	if ee != nil {
		return 0
	}

	return cc
}

func getClientCount(conn redis.Conn, vvv bool) int64 {
	var keys, err = redis.Strings(conn.Do("KEYS", "websocket_clients_count_*"))

	if err != nil {
		fmt.Println(0)
	}

	var count int64

	for idx := range keys {
		var key = keys[idx]
		var cc = doGetClientCount(conn, keys[idx])

		if vvv {
			fmt.Println(key, cc)
		}

		count = count + cc
	}

	return count
}

func doReport(count int64) {
	var client = &http.Client{}
	var json = "{\"time\":\"" + time.Now().String() + "\", \"count\":" + strconv.FormatInt(count, 10) + "}"
	req, err := http.NewRequest("POST", "http://localhost:9090", bytes.NewBufferString(json))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		bodystr := string(body)
		fmt.Println(bodystr)
	}
}

func main() {
	var args = os.Args[1:]
	var vvv = len(args) > 0 && args[0] == "-v"
	var dev = len(args) > 1 && args[1] == "-d"

	var redisURL = "redis://192.168.1.152:6379"

	if dev {
		redisURL = "redis://192.168.200.50:6379"
	}

	var conn = getRedisConnect(redisURL)
	defer conn.Close()

start:

	doReport(getClientCount(conn, vvv))
	time.Sleep(time.Second)

	goto start

}
