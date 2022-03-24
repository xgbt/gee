package geeWeb

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		// 启动计时器
		t := time.Now()
		// 执行请求
		c.Next()
		// 计算执行时间
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
