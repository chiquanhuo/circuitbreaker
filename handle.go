package circuit

import (
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"time"
	"testing"
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				fmt.Println("iwwwdial" + err.Error())
			}
			return c, err
		},
	}
}

var (
	pool        *redis.Pool
	redisServer = flag.String("redisServer", ":8889", "")
)

func TestBreaker(t *testing.T) {
	flag.Parse()
	pool = newPool(*redisServer)
	// breaker listen
	breaker := NewBreaker(0.1, 20, 3, time.Duration(5*time.Second))
	breaker.Subscribe()

	r := gin.Default()
	r.GET("/price/", func(c *gin.Context) {
		if breaker.GetStatus() == false {
			c.JSON(200, gin.H{
				"message": "false",
			})
			return
		}

		conn := pool.Get()
		defer conn.Close()
		_, err := conn.Do("SETEX", "testBreaker", 3600, "1")
		if err != nil {
			breaker.Call(false)
		} else {
			breaker.Call(true)
		}
		c.JSON(200, gin.H{
			"message": "test",
		})
		return
	})
	r.Run()
}
