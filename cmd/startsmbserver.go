package main

// StartSmbServer Main entry point to start the Simple Message Bus server with
// the given gateway configuration
// func StartSmbServer(gwConfig *config.GatewayConfig) (*smbserver.ServeSmbus, error) {
// 	var server *smbserver.ServeSmbus
// 	var err error

// 	// gwConfig, err := lib.SetupConfig(homeFolder, "", nil)

// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	if gwConfig.Messenger.CertFolder != "" {
// 		server, err = smbserver.StartTLS(gwConfig.Messenger.HostPort, gwConfig.Messenger.CertFolder)
// 	} else {
// 		server, err = smbserver.Start(gwConfig.Messenger.HostPort)
// 	}
// 	return server, err
// }
