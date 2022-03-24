package geeWeb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 用于存储JSON数据，在构建JSON数据时能更加简洁
type H map[string]interface{}

type Context struct {
	// 原始对象
	Writer http.ResponseWriter
	Req    *http.Request
	// 请求信息
	Method string
	Path   string
	Params map[string]string
	// 回应信息
	StatusCode int
	// 中间件数组及其下标idx
	handlers []HandlerFunc
	idx      int
	// engine 指针
	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Path:   req.URL.Path,
		Method: req.Method,
		Req:    req,
		Writer: w,
		idx:    -1,
	}
}

func (c *Context) Next() {
	c.idx++
	for ; c.idx < len(c.handlers); c.idx++ {
		c.handlers[c.idx](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.idx = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (c *Context) Param(key string) string {
	val, _ := c.Params[key]
	return val
}

// Query 查询Query参数
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// PostForm 查询PostForm参数
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// String 构造返回字符串的HTTP响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 构造返回JSON的HTTP响应
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Data 构造返回字符数组的HTTP响应
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 构造返回HTML的HTTP响应
// HTML template render
// refer https://golang.org/pkg/html/template/
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
