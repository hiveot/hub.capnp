package svcconfig

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"path"
	"path/filepath"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/logging"
)

// LoadServiceConfig Load a configuration file from the config folder and applies commandline options.
// This invokes Fatal if the configuration file is invalid, or required but not found.
//
// This invokes flag.Parse(). Flag commandline options added are:
//
//			 -c configFile
//			 --home directory
//			 --certs directory
//			 --services directory
//			 --logs directory
//		  --loglevel info|warning
//			 --run directory
//
//			If a 'cfg' interface is provided, the configuration is loaded from file and parsed as yaml.
//
//			serviceName is used for the configuration file with the '.yaml' extension
//			required returns an error if the configuration file doesn't exist
//			cfg is the interface to the configuration object. nil to ignore configuration and just load the folders.
//	 Returns the folder, service TLS certificate, and CA Certificate if found
//func LoadServiceConfig(
//	serviceName string, required bool, cfg interface{},
//) (f AppFolders, svcCert *tls.Certificate, caCert *x509.Certificate) {
//
//	// run the commandline options
//	var err error
//	var cfgData []byte
//	var certsFolder = ""
//	var homeFolder = ""
//	var logLevel = "info"
//	var logsFolder = ""
//	var runFolder = ""
//	var servicesFolder = ""
//	var storesFolder = ""
//
//	f = GetFolders(homeFolder, false)
//	cfgFile := path.Join(f.Config, serviceName+".yaml")
//	overrideCfgFile := ""
//	if cfg != nil {
//		flag.StringVar(&overrideCfgFile, "c", cfgFile, "Service config file")
//	}
//	flag.StringVar(&homeFolder, "home", f.Home, "Application home directory")
//	flag.StringVar(&certsFolder, "certs", f.Certs, "Certificates directory")
//	flag.StringVar(&servicesFolder, "services", f.Services, "Application services directory")
//	flag.StringVar(&logsFolder, "logs", f.Logs, "Service log files directory")
//	flag.StringVar(&logLevel, "loglevel", logLevel, "Loglevel info|warning. Default is warning")
//	flag.StringVar(&runFolder, "run", f.Run, "Runtime directory for sockets and pid files")
//	flag.StringVar(&storesFolder, "stores", f.Stores, "Storage directory")
//	flag.Parse()
//
//	// homefolder is special as it overrides all other folders
//	// detect the override by comparing original folder with assigned folder
//	f2 := GetFolders(homeFolder, false)
//	if certsFolder != f.Certs {
//		f2.Certs = certsFolder
//	}
//	if overrideCfgFile != cfgFile {
//		cfgFile = overrideCfgFile
//	} else if f2.Config != f.Config {
//		cfgFile = path.Join(f2.Config, serviceName+".yaml")
//	}
//	if servicesFolder != f.Services {
//		f2.Services = servicesFolder
//	}
//	if logsFolder != f.Logs {
//		f2.Logs = logsFolder
//	}
//	if runFolder != f.Run {
//		f2.Run = runFolder
//	}
//	f2.SocketPath = filepath.Join(f2.Run, serviceName+".socket")
//
//	if storesFolder != f.Stores {
//		f2.Stores = storesFolder
//	}
//	if logsFolder != "" {
//		logFile := path.Join(logsFolder, serviceName+".log")
//		logging.SetLogging(logLevel, logFile)
//	} else {
//		logging.SetLogging(logLevel, "")
//	}
//
//	// ignore configuration file if no destination interface is given
//	if cfg != nil {
//		cfgData, err = os.ReadFile(cfgFile)
//		if err == nil {
//			logrus.Infof("Loaded configuration file: %s", cfgFile)
//			err = yaml.Unmarshal(cfgData, cfg)
//			if err != nil {
//				logrus.Fatalf("Loading configuration file '%s' failed with: %s", cfgFile, err)
//			}
//		} else if !required {
//			logrus.Infof("Configuration file '%s' not found. Ignored.", cfgFile)
//			err = nil
//		} else {
//			logrus.Fatalf("Configuration file '%s' not found but required.", cfgFile)
//		}
//	}
//
//	// load the certificates if available
//	caCertPath := path.Join(f2.Certs, hubapi.DefaultCaCertFile)
//	caCert, _ = certsclient.LoadX509CertFromPEM(caCertPath)
//	svcCertPath := path.Join(f2.Certs, serviceName+"Cert.pem")
//	svcKeyPath := path.Join(f2.Certs, serviceName+"Key.pem")
//	svcCert, _ = certsclient.LoadTLSCertFromPEM(svcCertPath, svcKeyPath)
//
//	return f2, svcCert, caCert
//}

// SetupFolderConfig creates a folder configuration for based on commandline options.
// Returns the folders, service TLS certificate, and CA Certificate if found
//
// This invokes flag.Parse(). Flag commandline options added are:
//
//		 -c configFile
//		 --home directory
//		 --certs directory
//		 --services directory
//		 --logs directory
//	  --loglevel info|warning
//		 --run directory
//
//		If a 'cfg' interface is provided, the configuration is loaded from file and parsed as yaml.
//
//		serviceName is used for the configuration file with the '.yaml' extension
func SetupFolderConfig(serviceName string) (f AppFolders, svcCert *tls.Certificate, caCert *x509.Certificate) {

	// run the commandline options
	var certsFolder = ""
	var configFolder = ""
	var homeFolder = ""
	var logLevel = "info"
	var logsFolder = ""
	var runFolder = ""
	var servicesFolder = ""
	var storesFolder = ""
	var cfgFile = ""

	f = GetFolders(homeFolder, false)
	flag.StringVar(&homeFolder, "home", f.Home, "Application home directory")
	flag.StringVar(&certsFolder, "certs", f.Certs, "Certificates directory")
	flag.StringVar(&configFolder, "config", f.Config, "Configuration directory")
	flag.StringVar(&cfgFile, "c", "", "Service config file")
	flag.StringVar(&servicesFolder, "services", f.Services, "Application services directory")
	flag.StringVar(&logsFolder, "logs", f.Logs, "Service log files directory")
	flag.StringVar(&logLevel, "loglevel", logLevel, "Loglevel info|warning. Default is warning")
	flag.StringVar(&runFolder, "run", f.Run, "Runtime directory for sockets and pid files")
	flag.StringVar(&storesFolder, "stores", f.Stores, "Storage directory")
	flag.Parse()

	// homefolder is special as it overrides all default folders
	// detect the override by comparing original folder with assigned folder
	f2 := GetFolders(homeFolder, false)
	if certsFolder != f.Certs {
		f2.Certs = certsFolder
	}
	if configFolder != f.Config {
		f2.Config = configFolder
	}
	if cfgFile != "" {
		f2.ConfigFile = cfgFile
	} else {
		f2.ConfigFile = path.Join(f2.Config, serviceName+".yaml")
	}
	if servicesFolder != f.Services {
		f2.Services = servicesFolder
	}
	if logsFolder != f.Logs {
		f2.Logs = logsFolder
	}
	if runFolder != f.Run {
		f2.Run = runFolder
	}
	f2.SocketPath = filepath.Join(f2.Run, serviceName+".socket")

	if storesFolder != f.Stores {
		f2.Stores = storesFolder
	}
	if logsFolder != "" {
		logFile := path.Join(logsFolder, serviceName+".log")
		logging.SetLogging(logLevel, logFile)
	} else {
		logging.SetLogging(logLevel, "")
	}

	// load the certificates if available
	caCertPath := path.Join(f2.Certs, hubapi.DefaultCaCertFile)
	caCert, _ = certsclient.LoadX509CertFromPEM(caCertPath)
	svcCertPath := path.Join(f2.Certs, serviceName+"Cert.pem")
	svcKeyPath := path.Join(f2.Certs, serviceName+"Key.pem")
	svcCert, _ = certsclient.LoadTLSCertFromPEM(svcCertPath, svcKeyPath)

	return f2, svcCert, caCert
}
