package mosquittomgr

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// Generate a mosquitto.conf configuration file from template containing the
// ports and template from the plugin config.
// If a file exists it is replaced
//  hubConfig with the network host/port configuration
//  templateFilename filename of configuration template
//  configName  filename of the mosquitto configuration file to generate
//  Returns the final configuration path or an error
func ConfigureMosquitto(hubConfig *hubconfig.HubConfig, templateFilename string, configFilename string) (string, error) {

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
		"homeFolder":     hubConfig.Home,
		"logFolder":      path.Dir(hubConfig.LogFolder),
		"mqttCertPort":   fmt.Sprint(hubConfig.MqttCertPort),
		"mqttUnpwPortWS": fmt.Sprint(hubConfig.MqttUnpwPortWS),
	}
	configTxt := hubconfig.SubstituteText(string(configTemplate), templateParams)

	// write the configuration file
	if !path.IsAbs(configFilename) {
		configFilename = path.Join(hubConfig.ConfigFolder, configFilename)
	}
	ioutil.WriteFile(configFilename, []byte(configTxt), 0644)
	return configFilename, nil
}
