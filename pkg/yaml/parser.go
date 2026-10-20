package yaml

import (
	"bytes"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Settings    map[string]interface{} `yaml:"settings"`
	Database    DatabaseConfig         `yaml:"database"`
	Integrations []IntegrationConfig   `yaml:"integrations"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type IntegrationConfig struct {
	Name     string            `yaml:"name"`
	Type     string            `yaml:"type"`
	Endpoint string            `yaml:"endpoint"`
	Options  map[string]string `yaml:"options"`
}

func ParseConfig(data []byte) (*Config, error) {
	var config Config
	err := yaml.Unmarshal(data, &config)
	return &config, err
}

func ParseConfigFromReader(reader io.Reader) (*Config, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ParseConfig(data)
}

func ParseYAMLAny(data []byte) (interface{}, error) {
	var result interface{}
	err := yaml.Unmarshal(data, &result)
	return result, err
}

func ParseMultipleDocuments(data []byte) ([]interface{}, error) {
	var documents []interface{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var doc interface{}
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func ParseToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := yaml.Unmarshal(data, &result)
	return result, err
}

type YAMLProcessor struct {
	data []byte
}

func NewYAMLProcessor(data []byte) *YAMLProcessor {
	return &YAMLProcessor{data: data}
}

func (p *YAMLProcessor) Parse() (*Config, error) {
	return ParseConfig(p.data)
}

func (p *YAMLProcessor) ParseAny() (interface{}, error) {
	return ParseYAMLAny(p.data)
}

func (p *YAMLProcessor) GetValue(key string) (interface{}, error) {
	m, err := ParseToMap(p.data)
	if err != nil {
		return nil, err
	}
	return m[key], nil
}

func MergeConfigs(configs ...[]byte) (*Config, error) {
	merged := make(map[string]interface{})

	for _, data := range configs {
		var m map[string]interface{}
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, err
		}

		for k, v := range m {
			merged[k] = v
		}
	}

	mergedData, err := yaml.Marshal(merged)
	if err != nil {
		return nil, err
	}

	return ParseConfig(mergedData)
}

func ParseConfigBounded(data []byte, maxSize int64) (*Config, error) {
	if int64(len(data)) > maxSize {
		return nil, io.ErrShortBuffer
	}

	var config Config
	err := yaml.Unmarshal(data, &config)
	return &config, err
}

func SerializeConfig(config *Config) ([]byte, error) {
	return yaml.Marshal(config)
}

func ValidateConfig(data []byte) error {
	var config Config
	return yaml.Unmarshal(data, &config)
}
