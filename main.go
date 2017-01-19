package main

import "fmt"
import "github.com/garyburd/redigo/redis"
import "strconv"
import "os"

//获取一个redis连接
func getRedisConnect(redisURL string) redis.Conn {
	conn, err := redis.DialURL(redisURL)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return conn
}

//取客户端总数
func clientCount(c redis.Conn, key string) int64 {
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

	var keys, err = redis.Strings(conn.Do("KEYS", "websocket_clients_count_*"))

	if err != nil {
		fmt.Println(0)
		return
	}

	var count int64
	for idx := range keys {
		var key = keys[idx]
		var cc = clientCount(conn, keys[idx])

		if vvv {
			fmt.Println(key, cc)
		}

		count = count + cc
	}

	fmt.Println(count)
}
