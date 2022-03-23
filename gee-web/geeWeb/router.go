package geeWeb

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc), // 路由映射表
	}
}

// Only one * is allowed
func parsePattern(pattern string) []string {
	parts := make([]string, 0)
	for _, item := range strings.Split(pattern, "/") {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = new(node)
	}
	parts := parsePattern(pattern)
	r.roots[method].insert(pattern, parts, 0)
	// key 由请求方法和路由地址构成
	key := method + "-" + pattern
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, pattern string) (*node, map[string]string) {
	// 获取Method根节点
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	searchParts := parsePattern(pattern)
	n := root.search(searchParts, 0)
	if n != nil {
		params := make(map[string]string)
		parts := parsePattern(n.pattern)
		for idx, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[idx]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[idx:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

func (r *router) getRouters(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}

	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)

	if n != nil {
		key := c.Method + "-" + n.pattern
		c.Params = params
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}

	c.Next()
}
