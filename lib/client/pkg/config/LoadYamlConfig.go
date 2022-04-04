package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// LoadYamlConfig loads a yaml configuration from file and substitutes keywords variables
// with values from the provided map.
//
// This function doesn't really care what a keyword looks like but as convention "${name}"
// is used. If the keyword is not present in the given map then it will remain unchanged in
// the configFile.
//
//  configFile path to yaml configuration file
//  config interface to typed structure matching the config. Must have yaml tags
//  substituteMap map to substitude keys with value from map, nil to ignore
// Returns nil if successful
func LoadYamlConfig(configFile string, config interface{}, substituteMap map[string]string) error {
	var err error
	var rawConfig []byte

	if config == nil {
		err := fmt.Errorf("LoadYamlConfig: Loading config file: %s, but supplied configuration struct is nil", configFile)
		logrus.Error(err)
		return err
	}

	rawConfig, err = ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Infof("LoadYamlConfig: Error loading config file: %s: %s", configFile, err)
		return err
	}
	logrus.Infof("Loaded config file '%s'", configFile)
	rawText := string(rawConfig)
	if substituteMap != nil {
		rawText = SubstituteText(rawText, substituteMap)
	}

	err = yaml.Unmarshal([]byte(rawText), config)
	if err != nil {
		logrus.Errorf("LoadYamlConfig: Error parsing config file '%s': %s", configFile, err)
		return err
	}
	return nil
}

// SubstituteText substitutes template strings in the text
// This substitutes configuration values in the format ${name} with the given
// variable from the map. This format is compatible with JS templates.
// This is kept simple by requiring the full template keyword in the replacement
// map, eg use "${name}" and not "name" as the key in the map.
//
//  text to substitude template strings, eg "hello ${name}"
//  substituteMap with replacement keywords, eg {"${destination}":"world"}
// Returns text with template strings replaced
func SubstituteText(text string, substituteMap map[string]string) string {
	newText := text

	for key, val := range substituteMap {
		newText = strings.ReplaceAll(newText, key, val)
	}
	return newText
}
