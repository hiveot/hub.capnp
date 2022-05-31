package internal

import (
	"fmt"
	"io/ioutil"

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
//  configFolder with the network host/port configuration
//  templateFilename full filename of configuration template input file
//  configFileName  full filename of the mosquitto configuration output file
//  Returns nil or an error
func ConfigureMosquitto(mmConfig *MMConfig, templateFilename string, configFilename string) error {

	configTemplate, err := ioutil.ReadFile(templateFilename)
	if err != nil {
		logrus.Errorf("Unable to generate mosquitto configuration. Template file %s read error: %s", templateFilename, err)
		return err
	}
	templateParams := map[string]string{}
	templateParams["${mqttPortWS}"] = fmt.Sprint(mmConfig.MqttPortWS)
	templateParams["${mqttPortUnpw}"] = fmt.Sprint(mmConfig.MqttPortUnpw)
	templateParams["${mqttPortCert}"] = fmt.Sprint(mmConfig.MqttPortCert)
	templateParams["${caCertFile}"] = mmConfig.CaCertFile
	templateParams["${serverCertFile}"] = mmConfig.ServerCertFile
	templateParams["${serverKeyFile}"] = mmConfig.ServerKeyFile
	templateParams["${mosqAuthPlugin}"] = mmConfig.MosqAuthPlugin
	templateParams["${aclFile}"] = mmConfig.AclFile
	templateParams["${logFolder}"] = mmConfig.LogFolder

	//for k, v := range mmConfig.AsMap() {
	//	name := "${" + k + "}"
	//	templateParams[name] = v
	//}
	configTxt := config.SubstituteText(string(configTemplate), templateParams)

	// write the configuration file
	ioutil.WriteFile(configFilename, []byte(configTxt), 0644)
	return nil
}
