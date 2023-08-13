package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Sections []*SectionConfig
	Template string

	sectionIndex map[string]*SectionConfig
}

type SectionConfig struct {
	Name         string
	Rules        []*SectionRule
	Order        int
	Skip         bool
	SkipPerMonth bool `yaml:"skipPerMonth"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %q: %w", path, err)
	}
	defer file.Close()

	dec := yaml.NewDecoder(file)
	cfg := &Config{
		sectionIndex: make(map[string]*SectionConfig),
	}
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("yaml decode %q: %w", path, err)
	}

	if err := cfg.compile(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (self *Config) compile() error {
	for _, sect := range self.Sections {
		self.sectionIndex[sect.Name] = sect
		for i, rule := range sect.Rules {
			if err := rule.compile(sect.Name, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *Config) FindSection(rec Record) (string, string, error) {
	for _, sect := range self.Sections {
		for _, rule := range sect.Rules {
			if key, err := rule.ExtractKey(rec); err != nil {
				return "", "", err
			} else if key != "" {
				return sect.Name, key, nil
			}
		}
	}
	return "", "", nil
}

func (self *Config) SkipFromSum(sectName string) bool {
	sect := self.sectionIndex[sectName]
	return sect.Skip
}

func (self *Config) SkipPerMonth(sectName string) bool {
	sect := self.sectionIndex[sectName]
	return sect.SkipPerMonth
}
