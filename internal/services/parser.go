package services

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"

	"github.com/buger/jsonparser"
	"github.com/gogo/protobuf/proto"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

func ParseYAMLConfig(data []byte) (map[string]interface{}, error) {
	var config map[string]interface{}
	err := yaml.Unmarshal(data, &config)
	return config, err
}

func ParseYAMLStrict(data []byte) (interface{}, error) {
	var result interface{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&result)
	return result, err
}

func ParseJSONWithGjson(data []byte, path string) string {
	return gjson.GetBytes(data, path).String()
}

func ParseJSONDeep(data []byte, keys ...string) ([]byte, error) {
	val, _, _, err := jsonparser.Get(data, keys...)
	return val, err
}

func ParseJSONArray(data []byte, key string, callback func([]byte) error) error {
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		callback(value)
	}, key)
	return err
}

type XMLDocument struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
}

func ParseXMLDocument(data []byte) (*XMLDocument, error) {
	var doc XMLDocument
	err := xml.Unmarshal(data, &doc)
	return &doc, err
}

func ParseXMLWithDecoder(reader io.Reader) (*XMLDocument, error) {
	decoder := xml.NewDecoder(reader)
	var doc XMLDocument
	err := decoder.Decode(&doc)
	return &doc, err
}

func ParseProtobuf(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

type ConfigParser struct {
	data []byte
}

func NewConfigParser(data []byte) *ConfigParser {
	return &ConfigParser{data: data}
}

func (p *ConfigParser) ParseAsYAML() (map[string]interface{}, error) {
	return ParseYAMLConfig(p.data)
}

func (p *ConfigParser) GetJSONValue(path string) string {
	return ParseJSONWithGjson(p.data, path)
}

func ValidateRegexPattern(pattern, input string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(input), nil
}

func MatchWithPattern(pattern, input string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.FindAllString(input, -1), nil
}

func ParseYAMLBounded(data []byte, maxSize int) (map[string]interface{}, error) {
	if len(data) > maxSize {
		return nil, io.ErrShortBuffer
	}

	var config map[string]interface{}
	err := yaml.Unmarshal(data, &config)
	return config, err
}

func ParseXMLNoEntities(data []byte) (*XMLDocument, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Entity = make(map[string]string)

	var doc XMLDocument
	err := decoder.Decode(&doc)
	return &doc, err
}

func ValidateRegexBounded(pattern string) bool {
	if len(pattern) > 1000 {
		return false
	}

	dangerous := []string{"(.+)+", "(.*)*", "([a-zA-Z]+)*"}
	for _, d := range dangerous {
		if pattern == d {
			return false
		}
	}

	_, err := regexp.Compile(pattern)
	return err == nil
}

type DataProcessor interface {
	Process(data []byte) (interface{}, error)
}

type YAMLProcessor struct{}

func (p *YAMLProcessor) Process(data []byte) (interface{}, error) {
	return ParseYAMLConfig(data)
}

type JSONProcessor struct {
	path string
}

func (p *JSONProcessor) Process(data []byte) (interface{}, error) {
	return ParseJSONWithGjson(data, p.path), nil
}

func ProcessWithHandler(data []byte, processor DataProcessor) (interface{}, error) {
	return processor.Process(data)
}
