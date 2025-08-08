package pipeline

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Pipeline 数据管道接口
type Pipeline interface {
	ProcessItem(item map[string]interface{}) map[string]interface{}
	Open() error
	Close() error
}

// ConsolePipeline 控制台输出管道
type ConsolePipeline struct{}

// NewConsolePipeline 创建控制台管道
func NewConsolePipeline() *ConsolePipeline {
	return &ConsolePipeline{}
}

// ProcessItem 处理数据项
func (p *ConsolePipeline) ProcessItem(item map[string]interface{}) map[string]interface{} {
	fmt.Printf("Item: %+v\n", item)
	return item
}

// Open 打开管道
func (p *ConsolePipeline) Open() error {
	return nil
}

// Close 关闭管道
func (p *ConsolePipeline) Close() error {
	return nil
}

// JSONPipeline JSON文件管道
type JSONPipeline struct {
	filename string
	file     *os.File
	encoder  *json.Encoder
	mutex    sync.Mutex
}

// NewJSONPipeline 创建JSON管道
func NewJSONPipeline(filename string) *JSONPipeline {
	return &JSONPipeline{
		filename: filename,
	}
}

// ProcessItem 处理数据项
func (p *JSONPipeline) ProcessItem(item map[string]interface{}) map[string]interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.encoder != nil {
		p.encoder.Encode(item)
	}
	
	return item
}

// Open 打开管道
func (p *JSONPipeline) Open() error {
	// 确保目录存在
	dir := filepath.Dir(p.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}
	
	file, err := os.Create(p.filename)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	
	p.file = file
	p.encoder = json.NewEncoder(file)
	p.encoder.SetIndent("", "  ")
	
	return nil
}

// Close 关闭管道
func (p *JSONPipeline) Close() error {
	if p.file != nil {
		return p.file.Close()
	}
	return nil
}

// CSVPipeline CSV文件管道
type CSVPipeline struct {
	filename   string
	file       *os.File
	writer     *csv.Writer
	headers    []string
	headerSet  bool
	mutex      sync.Mutex
}

// NewCSVPipeline 创建CSV管道
func NewCSVPipeline(filename string, headers []string) *CSVPipeline {
	return &CSVPipeline{
		filename: filename,
		headers:  headers,
	}
}

// ProcessItem 处理数据项
func (p *CSVPipeline) ProcessItem(item map[string]interface{}) map[string]interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.writer == nil {
		return item
	}
	
	// 写入表头
	if !p.headerSet && len(p.headers) > 0 {
		p.writer.Write(p.headers)
		p.headerSet = true
	}
	
	// 写入数据行
	record := make([]string, len(p.headers))
	for i, header := range p.headers {
		if value, exists := item[header]; exists {
			record[i] = fmt.Sprintf("%v", value)
		}
	}
	
	p.writer.Write(record)
	p.writer.Flush()
	
	return item
}

// Open 打开管道
func (p *CSVPipeline) Open() error {
	// 确保目录存在
	dir := filepath.Dir(p.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}
	
	file, err := os.Create(p.filename)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	
	p.file = file
	p.writer = csv.NewWriter(file)
	
	return nil
}

// Close 关闭管道
func (p *CSVPipeline) Close() error {
	if p.writer != nil {
		p.writer.Flush()
	}
	if p.file != nil {
		return p.file.Close()
	}
	return nil
}

// XMLPipeline XML文件管道
type XMLPipeline struct {
	filename string
	file     *os.File
	encoder  *xml.Encoder
	mutex    sync.Mutex
	rootName string
}

// XMLItem XML项目包装
type XMLItem struct {
	XMLName xml.Name               `xml:"item"`
	Data    map[string]interface{} `xml:",any"`
}

// NewXMLPipeline 创建XML管道
func NewXMLPipeline(filename, rootName string) *XMLPipeline {
	if rootName == "" {
		rootName = "items"
	}
	
	return &XMLPipeline{
		filename: filename,
		rootName: rootName,
	}
}

// ProcessItem 处理数据项
func (p *XMLPipeline) ProcessItem(item map[string]interface{}) map[string]interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.encoder != nil {
		xmlItem := XMLItem{Data: item}
		p.encoder.Encode(xmlItem)
	}
	
	return item
}

// Open 打开管道
func (p *XMLPipeline) Open() error {
	// 确保目录存在
	dir := filepath.Dir(p.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}
	
	file, err := os.Create(p.filename)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	
	p.file = file
	p.encoder = xml.NewEncoder(file)
	p.encoder.Indent("", "  ")
	
	// 写入XML头和根元素开始标签
	file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	file.WriteString(fmt.Sprintf("<%s>\n", p.rootName))
	
	return nil
}

// Close 关闭管道
func (p *XMLPipeline) Close() error {
	if p.file != nil {
		// 写入根元素结束标签
		p.file.WriteString(fmt.Sprintf("</%s>\n", p.rootName))
		return p.file.Close()
	}
	return nil
}

// FilterPipeline 过滤管道
type FilterPipeline struct {
	filterFunc func(map[string]interface{}) bool
}

// NewFilterPipeline 创建过滤管道
func NewFilterPipeline(filterFunc func(map[string]interface{}) bool) *FilterPipeline {
	return &FilterPipeline{
		filterFunc: filterFunc,
	}
}

// ProcessItem 处理数据项
func (p *FilterPipeline) ProcessItem(item map[string]interface{}) map[string]interface{} {
	if p.filterFunc != nil && !p.filterFunc(item) {
		return nil // 过滤掉该项
	}
	return item
}

// Open 打开管道
func (p *FilterPipeline) Open() error {
	return nil
}

// Close 关闭管道
func (p *FilterPipeline) Close() error {
	return nil
}

// TransformPipeline 转换管道
type TransformPipeline struct {
	transformFunc func(map[string]interface{}) map[string]interface{}
}

// NewTransformPipeline 创建转换管道
func NewTransformPipeline(transformFunc func(map[string]interface{}) map[string]interface{}) *TransformPipeline {
	return &TransformPipeline{
		transformFunc: transformFunc,
	}
}

// ProcessItem 处理数据项
func (p *TransformPipeline) ProcessItem(item map[string]interface{}) map[string]interface{} {
	if p.transformFunc != nil {
		return p.transformFunc(item)
	}
	return item
}

// Open 打开管道
func (p *TransformPipeline) Open() error {
	return nil
}

// Close 关闭管道
func (p *TransformPipeline) Close() error {
	return nil
}