package geeWeb

import (
	"fmt"
	"strings"
)

type node struct {
	pattern  string  // 节点待匹配路由, 例如 /p/:lang
	part     string  // 节点存储的部分路由，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool    // 标记节点是否与 * 和 : 匹配, part 含有 : 或 * 时为true
}

func (n *node) String() string {
	return fmt.Sprintf("node[pattern=%s, part=%s, isWild=%t}", n.pattern, n.part, n.isWild)
}

// 返回第一个匹配成功的节点，用于insert操作
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 返回所有匹配成功的节点，用于search
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) insert(pattern string, parts []string, idx int) {
	if idx == len(parts) {
		n.pattern = pattern
		return
	}

	part := parts[idx]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, idx+1)
}

func (n *node) search(parts []string, idx int) *node {
	if idx == len(parts) || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[idx]
	children := n.matchChildren(part)
	for _, child := range children {
		ret := child.search(parts, idx+1)
		if ret != nil {
			return ret
		}
	}

	return nil
}

func (n *node) travel(list *[]*node) {
	if n.pattern != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}
