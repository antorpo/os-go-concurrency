package config

import (
	"encoding/json"
	"sync"

	cfg "github.com/antorpo/os-go-concurrency/internal/domain/entities/config"
	toolkit "github.com/antorpo/os-go-concurrency/pkg/config"
)

type configuration struct {
	mutex  sync.RWMutex
	config *cfg.Configuration
}

type IConfiguration interface {
	GetConfig() *cfg.Configuration
	LoadConfig() error
	LoadJSONProfile(profileName string, mappingType interface{}) (interface{}, error)
}

func NewConfiguration() IConfiguration {
	return &configuration{}
}

func (c *configuration) LoadConfig() error {
	var config cfg.Configuration

	if _, err := c.LoadJSONProfile(cfg.AppConfigName, &config.App); err != nil {
		return err
	}

	c.setConfig(&config)
	return nil
}

func (c *configuration) LoadJSONProfile(profileName string, mappingType interface{}) (interface{}, error) {
	bytes, err := readProfile(profileName)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &mappingType)
	if err != nil {
		return nil, err
	}

	return &mappingType, nil
}

func (c *configuration) GetConfig() *cfg.Configuration {
	if c.config == nil {
		return &cfg.Configuration{}
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.config
}

func (c *configuration) setConfig(newConfig *cfg.Configuration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.config = newConfig
}

func readProfile(profileName string) ([]byte, error) {
	return toolkit.Read(profileName)
}
