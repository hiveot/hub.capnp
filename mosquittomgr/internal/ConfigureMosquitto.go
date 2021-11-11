package internal

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/config"
)

// ConfigureMosquitto generates a mosquitto.conf configuration file from template containing the
// ports and template from the plugin config.
// If a file exists it is replaced
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
	// TODO: template keywords. These names MUST match hub.yaml :/
	templateParams := map[string]string{
		"${appFolder}":    hubConfig.AppFolder,
		"${binFolder}":    path.Join(hubConfig.AppFolder, "bin"),
		"${logsFolder}":   hubConfig.LogsFolder,
		"${configFolder}": hubConfig.ConfigFolder,
		"${caCertFile}":   path.Join(hubConfig.CertsFolder, config.DefaultCaCertFile),
		"${hubCertFile}":  path.Join(hubConfig.CertsFolder, config.DefaultServerCertFile),
		"${hubKeyFile}":   path.Join(hubConfig.CertsFolder, config.DefaultServerKeyFile),
		"${mqttPortCert}": fmt.Sprint(hubConfig.MqttPortCert),
		"${mqttPortUnpw}": fmt.Sprint(hubConfig.MqttPortUnpw),
		"${mqttPortWS}":   fmt.Sprint(hubConfig.MqttPortWS),
	}
	configTxt := config.SubstituteText(string(configTemplate), templateParams)

	// write the configuration file
	if !path.IsAbs(configFilename) {
		configFilename = path.Join(hubConfig.ConfigFolder, configFilename)
	}
	ioutil.WriteFile(configFilename, []byte(configTxt), 0644)
	return configFilename, nil
}
