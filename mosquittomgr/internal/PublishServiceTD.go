package internal

// PublishServiceTD generates a TD for this service and publishes it.
// It includes the admin HTTPS API to manage mosquitto.
// func (mm *MosquittoManager) PublishServiceTD() {
// 	serviceID := mm.Config.ClientID
// 	deviceType := vocab.DeviceTypeService
// 	thingID := td.CreateThingID(mm.hubConfig.Zone, serviceID, deviceType)
// 	thingTD := td.CreateTD(thingID, deviceType)
// 	td.SetThingDescription(thingTD, "Mosquitto admin", "Administration of mosquitto server through CLI")

// 	// Add CLI service hostname, IP and port
// 	prop := td.CreateProperty(vocab.PropNameAddress, "Admin CLI address", vocab.PropertyTypeAttr)
// 	td.SetPropertyValue(prop, mm.Config.CLIAddress)
// 	td.AddTDProperty(thingTD, vocab.PropNameAddress, prop)

// 	prop = td.CreateProperty(vocab.PropNameHostname, "Admin CLI hostname", vocab.PropertyTypeAttr)
// 	td.SetPropertyValue(prop, mm.Config.CLIHost)
// 	td.AddTDProperty(thingTD, vocab.PropNameHostname, prop)

// 	prop = td.CreateProperty(vocab.PropNamePort, "Admin CLI port nr", vocab.PropertyTypeAttr)
// 	td.SetPropertyValue(prop, fmt.Sprint(mm.Config.CLIPort))
// 	td.AddTDProperty(thingTD, vocab.PropNamePort, prop)

// 	mm.hubClient.PublishTD(thingID, thingTD)
// }
