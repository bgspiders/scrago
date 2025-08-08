package selector

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Selector 选择器
type Selector struct {
	doc  *goquery.Document
	node *html.Node
	text string
}

// Selection 选择结果
type Selection struct {
	nodes []*html.Node
	text  []string
}

// NewSelector 创建新选择器
func NewSelector(text string) *Selector {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(text))
	if doc == nil {
		// 如果HTML解析失败，创建空选择器
		return &Selector{text: text}
	}
	
	return &Selector{
		doc:  doc,
		text: text,
	}
}

// CSS 使用CSS选择器
func (s *Selector) CSS(cssSelector string) *Selection {
	if s.doc == nil {
		return &Selection{}
	}
	
	selection := s.doc.Find(cssSelector)
	nodes := make([]*html.Node, selection.Length())
	texts := make([]string, selection.Length())
	
	selection.Each(func(i int, sel *goquery.Selection) {
		if sel.Nodes != nil && len(sel.Nodes) > 0 {
			nodes[i] = sel.Nodes[0]
			texts[i] = sel.Text()
		}
	})
	
	return &Selection{
		nodes: nodes,
		text:  texts,
	}
}

// XPath 使用XPath选择器
func (s *Selector) XPath(xpathExpr string) *Selection {
	var nodes []*html.Node
	var texts []string
	
	if s.doc != nil {
		// 从goquery文档获取根节点
		if s.doc.Nodes != nil && len(s.doc.Nodes) > 0 {
			htmlNodes := htmlquery.Find(s.doc.Nodes[0], xpathExpr)
			for _, node := range htmlNodes {
				nodes = append(nodes, node)
				texts = append(texts, htmlquery.InnerText(node))
			}
		}
	}
	
	return &Selection{
		nodes: nodes,
		text:  texts,
	}
}

// Regex 使用正则表达式
func (s *Selector) Regex(pattern string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return []string{}
	}
	
	return re.FindAllString(s.text, -1)
}

// Find 查找子选择器
func (s *Selector) Find(cssSelector string) *Selection {
	return s.CSS(cssSelector)
}

// Text 获取文本内容
func (s *Selector) Text() string {
	if s.doc != nil {
		return s.doc.Text()
	}
	return s.text
}

// HTML 获取HTML内容
func (s *Selector) HTML() string {
	if s.doc != nil {
		html, _ := s.doc.Html()
		return html
	}
	return s.text
}

// Selection 方法

// Get 获取指定索引的元素
func (sel *Selection) Get(index int) *Selector {
	if index < 0 || index >= len(sel.nodes) {
		return &Selector{}
	}
	
	if sel.nodes[index] != nil {
		doc := &goquery.Document{Selection: &goquery.Selection{Nodes: []*html.Node{sel.nodes[index]}}}
		return &Selector{doc: doc}
	}
	
	return &Selector{}
}

// First 获取第一个元素
func (sel *Selection) First() *Selector {
	return sel.Get(0)
}

// Last 获取最后一个元素
func (sel *Selection) Last() *Selector {
	return sel.Get(len(sel.nodes) - 1)
}

// Length 获取元素数量
func (sel *Selection) Length() int {
	return len(sel.nodes)
}

// Text 获取所有元素的文本
func (sel *Selection) Text() string {
	if len(sel.text) > 0 {
		return strings.Join(sel.text, " ")
	}
	return ""
}

// Texts 获取所有元素的文本数组
func (sel *Selection) Texts() []string {
	return sel.text
}

// Attr 获取属性值
func (sel *Selection) Attr(attrName string) string {
	if len(sel.nodes) == 0 {
		return ""
	}
	
	doc := &goquery.Document{Selection: &goquery.Selection{Nodes: []*html.Node{sel.nodes[0]}}}
	attr, _ := doc.Selection.Attr(attrName)
	return attr
}

// Attrs 获取所有元素的属性值
func (sel *Selection) Attrs(attrName string) []string {
	attrs := make([]string, 0, len(sel.nodes))
	
	for _, node := range sel.nodes {
		if node != nil {
			doc := &goquery.Document{Selection: &goquery.Selection{Nodes: []*html.Node{node}}}
			if attr, exists := doc.Selection.Attr(attrName); exists {
				attrs = append(attrs, attr)
			}
		}
	}
	
	return attrs
}

// CSS 在Selection上使用CSS选择器
func (sel *Selection) CSS(cssSelector string) *Selection {
	allNodes := make([]*html.Node, 0)
	allTexts := make([]string, 0)
	
	for _, node := range sel.nodes {
		if node != nil {
			doc := &goquery.Document{Selection: &goquery.Selection{Nodes: []*html.Node{node}}}
			selection := doc.Find(cssSelector)
			
			selection.Each(func(i int, s *goquery.Selection) {
				if s.Nodes != nil && len(s.Nodes) > 0 {
					allNodes = append(allNodes, s.Nodes[0])
					allTexts = append(allTexts, s.Text())
				}
			})
		}
	}
	
	return &Selection{
		nodes: allNodes,
		text:  allTexts,
	}
}

// Each 遍历所有元素
func (sel *Selection) Each(fn func(int, *Selector)) {
	for i := range sel.nodes {
		fn(i, sel.Get(i))
	}
}

// Map 映射所有元素
func (sel *Selection) Map(fn func(int, *Selector) string) []string {
	results := make([]string, 0, len(sel.nodes))
	
	for i := range sel.nodes {
		result := fn(i, sel.Get(i))
		results = append(results, result)
	}
	
	return results
}

// Filter 过滤元素
func (sel *Selection) Filter(cssSelector string) *Selection {
	filteredNodes := make([]*html.Node, 0)
	filteredTexts := make([]string, 0)
	
	for i, node := range sel.nodes {
		if node != nil {
			doc := &goquery.Document{Selection: &goquery.Selection{Nodes: []*html.Node{node}}}
			if doc.Selection.Is(cssSelector) {
				filteredNodes = append(filteredNodes, node)
				if i < len(sel.text) {
					filteredTexts = append(filteredTexts, sel.text[i])
				}
			}
		}
	}
	
	return &Selection{
		nodes: filteredNodes,
		text:  filteredTexts,
	}
}