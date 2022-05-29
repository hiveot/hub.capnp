package internal

import (
	"io/ioutil"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wost-go/pkg/config"
)

// ConfigureMosquitto generates a mosquitto.conf configuration file.
//
// It supports the use of file, folder and certificate variables in the config template that are replaced
// with yaml encoded values from the provided HubConfig.
//
// For example ${configFolder} is replaced by the value of HubConfig's "ConfigFolder".
//
// If a file exists it is replaced.
//  hubConfig with the network host/port configuration
//  templateFilename filename of configuration template
//  configName  filename of the mosquitto configuration file to generate
//  Returns the final configuration path or an error
func ConfigureMosquitto(hubConfig *config.HubConfig, templateFilename string, configFilename string) (string, error) {

	// load the template
	if !path.IsAbs(templateFilename) {
		templateFilename = path.Join(hubConfig.ConfigFolder, templateFilename)
	}
	configTemplate, err := ioutil.ReadFile(templateFilename)
	if err != nil {
		logrus.Errorf("Unable to generate mosquitto configuration. Template file %s read error: %s", templateFilename, err)
		return "", err
	}
	// TODO: iterate HubConfig values
	templateParams := map[string]string{}
	// FIXME: server cert info is not included in hubconfig
	templateParams["${serverCertFile}"] = path.Join(hubConfig.CertsFolder, config.DefaultServerCertFile)
	templateParams["${serverKeyFile}"] = path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile)
	for k, v := range hubConfig.AsMap() {
		name := "${" + k + "}"
		templateParams[name] = v
	}
	configTxt := config.SubstituteText(string(configTemplate), templateParams)

	// write the configuration file
	if !path.IsAbs(configFilename) {
		configFilename = path.Join(hubConfig.ConfigFolder, configFilename)
	}
	ioutil.WriteFile(configFilename, []byte(configTxt), 0644)
	return configFilename, nil
}
