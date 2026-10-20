package xml

import (
	"bytes"
	"encoding/xml"
	"io"
	"io/ioutil"
)

type Document struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  string     `xml:",innerxml"`
	Children []Node     `xml:",any"`
}

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
}

type Configuration struct {
	XMLName  xml.Name          `xml:"configuration"`
	Settings []ConfigSetting   `xml:"setting"`
	Imports  []ConfigImport    `xml:"import"`
}

type ConfigSetting struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type ConfigImport struct {
	Path string `xml:"path,attr"`
	Type string `xml:"type,attr"`
}

func ParseXML(data []byte) (*Document, error) {
	var doc Document
	err := xml.Unmarshal(data, &doc)
	return &doc, err
}

func ParseXMLFromReader(reader io.Reader) (*Document, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ParseXML(data)
}

func ParseConfiguration(data []byte) (*Configuration, error) {
	var config Configuration
	err := xml.Unmarshal(data, &config)
	return &config, err
}

type XMLProcessor struct {
	data []byte
}

func NewXMLProcessor(data []byte) *XMLProcessor {
	return &XMLProcessor{data: data}
}

func (p *XMLProcessor) Parse() (*Document, error) {
	return ParseXML(p.data)
}

func (p *XMLProcessor) ParseConfig() (*Configuration, error) {
	return ParseConfiguration(p.data)
}

func ParseWithDecoder(reader io.Reader) (*Document, error) {
	decoder := xml.NewDecoder(reader)
	var doc Document
	err := decoder.Decode(&doc)
	return &doc, err
}

func ParseWithExternalEntities(data []byte) (*Document, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	var doc Document
	err := decoder.Decode(&doc)
	return &doc, err
}

type XMLHandler interface {
	Process(data []byte) (*Document, error)
}

type StandardHandler struct{}

func (h *StandardHandler) Process(data []byte) (*Document, error) {
	return ParseXML(data)
}

func ProcessWithHandler(handler XMLHandler, data []byte) (*Document, error) {
	return handler.Process(data)
}

func ParseXMLNoEntities(data []byte) (*Document, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Entity = make(map[string]string)

	var doc Document
	err := decoder.Decode(&doc)
	return &doc, err
}

func ParseConfigurationBounded(data []byte, maxSize int) (*Configuration, error) {
	if len(data) > maxSize {
		return nil, io.ErrShortBuffer
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Entity = make(map[string]string)

	var config Configuration
	err := decoder.Decode(&config)
	return &config, err
}

func SerializeToXML(v interface{}) ([]byte, error) {
	return xml.MarshalIndent(v, "", "  ")
}

func ValidateXMLStructure(data []byte) error {
	var doc Document
	return xml.Unmarshal(data, &doc)
}

type StreamProcessor struct {
	decoder *xml.Decoder
}

func NewStreamProcessor(reader io.Reader) *StreamProcessor {
	return &StreamProcessor{
		decoder: xml.NewDecoder(reader),
	}
}

func (p *StreamProcessor) NextElement() (*Node, error) {
	var node Node
	err := p.decoder.Decode(&node)
	return &node, err
}

func (p *StreamProcessor) ProcessAll(callback func(*Node) error) error {
	for {
		node, err := p.NextElement()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := callback(node); err != nil {
			return err
		}
	}
}
