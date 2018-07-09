package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v1"
)

type configStruct struct {
	DB struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		UserName string `yaml:"username"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	}
	Telegram struct {
		Token string `yaml:"token"`
	}
	Twitter struct {
		ConsumerKey    string `yaml:"consumerKey"`
		ConsumerSecret string `yaml:"consumerSecret"`
		Token          string `yaml:"token"`
		TokenSecret    string `yaml:"tokenSecret"`
	}
	Vk struct {
		SecureKey string `yaml:"secureKey"`
	}
	// List struct {
	// 	Rss     []string
	// 	Vk      []string
	// 	Twitter []string
	// }
}

func getConfig(configPath string) (config *configStruct, err error) {
	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
