// Package vocab with iotschema vocabulary for sensor, actuator and unitname names
// TODO: base this of a universally accepted ontology. Closest is iotschema.org but that seems incomplete
package vocab

// DataType of configuration, input and ouput values.
// type DataType string

// Available data types. See WoT vocabulary WoTDataTypeXxx
const (
// DataTypeArray value is an array of ?
// DataTypeArray DataType = "array"
// DataTypeBool value is true/false, 1/0, on/off
// DataTypeBool DataType = "boolean"
// DataTypeBytes value is encoded byte array
// DataTypeBytes DataType = "wost:bytes"
// DataTypeDate ISO8601 date YYYY-MM-DDTHH:MM:SS.mmmZ
// DataTypeDate DataType = "wost:date"
// DataTypeEnum value is one of a predefined set of string values, published in the 'enum info field'
// DataTypeEnum DataType = "wost:enum"
// DataTypeInt value is an integer number
// DataTypeInt DataType = "integer"
// value is a float number
// DataTypeNumber DataType = "number"
// a secret string that is not published
// value is an object with its own property definitions
// DataTypeObject DataType = "object"
// DataTypeSecret DataType = "wost:secret"
// DataTypeString DataType = "string"
// 3D vector (x, y, z) or (lat, lon, 0)
// DataTypeVector DataType = "wost:vector"
// value is a json object
// DataTypeJSON DataType = "wost:json"
)

// WoST attributes used in properties
const (
	AttrNameValue = "value"
)

// IoT Device types

// DeviceType identifying the purpose of the device
type DeviceType string

// Various Types of devices.
const (
	DeviceTypeAlarm          DeviceType = "alarm"          // an alarm emitter
	DeviceTypeAVControl      DeviceType = "avControl"      // Audio/Video controller
	DeviceTypeAVReceiver     DeviceType = "avReceiver"     // Node is a (not so) smart radio/receiver/amp (eg, denon)
	DeviceTypeBeacon         DeviceType = "beacon"         // device is a location beacon
	DeviceTypeButton         DeviceType = "button"         // device is a physical button device with one or more buttons
	DeviceTypeAdapter        DeviceType = "adapter"        // software adapter or service, eg virtual device
	DeviceTypePhone          DeviceType = "phone"          // device is a phone
	DeviceTypeCamera         DeviceType = "camera"         // Node with camera
	DeviceTypeComputer       DeviceType = "computer"       // General purpose computer
	DeviceTypeDimmer         DeviceType = "dimmer"         // light dimmer
	DeviceTypeGateway        DeviceType = "gateway"        // Node is a gateway for other nodes (onewire, zwave, etc)
	DeviceTypeKeypad         DeviceType = "keypad"         // Entry key pad
	DeviceTypeLock           DeviceType = "lock"           // Electronic door lock
	DeviceTypeMultisensor    DeviceType = "multisensor"    // Node with multiple sensors
	DeviceTypeNetRepeater    DeviceType = "netRepeater"    // Node is a zwave or other network repeater
	DeviceTypeNetRouter      DeviceType = "netRouter"      // Node is a network router
	DeviceTypeNetSwitch      DeviceType = "netSwitch"      // Node is a network switch
	DeviceTypeNetWifiAP      DeviceType = "wifiAP"         // Node is a wifi access point
	DeviceTypeOnOffSwitch    DeviceType = "onOffSwitch"    // Node is a physical on/off switch
	DeviceTypePowerMeter     DeviceType = "powerMeter"     // Node is a power meter
	DeviceTypeSensor         DeviceType = "sensor"         // Node is a single sensor (volt,...)
	DeviceTypeService        DeviceType = "service"        // Node provides a service
	DeviceTypeSmartlight     DeviceType = "smartlight"     // Node is a smart light, eg philips hue
	DeviceTypeThermometer    DeviceType = "thermometer"    // Node is a temperature meter
	DeviceTypeThermostat     DeviceType = "thermostat"     // Node is a thermostat control unit
	DeviceTypeTV             DeviceType = "tv"             // Node is a (not so) smart TV
	DeviceTypeUnknown        DeviceType = "unknown"        // type not identified
	DeviceTypeWallpaper      DeviceType = "wallpaper"      // Node is a wallpaper montage of multiple images
	DeviceTypeWaterValve     DeviceType = "waterValve"     // Water valve control unit
	DeviceTypeWeatherService DeviceType = "weatherService" // Node is a service providing current and forecasted weather
	DeviceTypeWeatherStation DeviceType = "weatherStation" // Node is a weatherstation device
	DeviceTypeWeighScale     DeviceType = "weighScale"     // Node is an electronic weight scale
)

// ThingPropType with types of property that are defined
type ThingPropType string

// Since property types is not part of the WoT vocabulary they are defined
// as part of the WoST vocabulary. Used in "@type" for TD Properties.
const (
	// Property is an actuator (readonly, use Actions)
	PropertyTypeActuator ThingPropType = "wost:actuator"
	// Property is a readonly internal Thing attribute
	PropertyTypeAttr ThingPropType = "wost:attr"
	// Property is a writable configuration
	PropertyTypeConfig ThingPropType = "wost:configuration"
	// Property is a readonly sensor
	PropertyTypeSensor ThingPropType = "wost:sensor"
	// Property is a readonly internal state
	PropertyTypeState ThingPropType = "wost:state"
	// Property is an input (use in Actions)
	PropertyTypeInput ThingPropType = "wost:input"
	// Property is an output (when different from sensor)
	PropertyTypeOutput ThingPropType = "wost:output"
)

// Vocabulary property names to be used by Things and plugins.
const (
	PropNameAcceleration           string = "acceleration"
	PropNameAddress                string = "address" // device domain or ip address
	PropNameAirQuality             string = "airquality"
	PropNameAlarm                  string = "alarm"
	PropNameAtmosphericPressure    string = "atmosphericpressure"
	PropNameBatch                  string = "batch" // Batch publishing size
	PropNameBattery                string = "battery"
	PropNameCarbonDioxideLevel     string = "co2level"
	PropNameCarbonMonoxideDetector string = "codetector"
	PropNameCarbonMonoxideLevel    string = "colevel"
	PropNameChannel                string = "avchannel"
	PropNameColor                  string = "color" // Color in hex notation
	PropNameColorTemperature       string = "colortemperature"
	PropNameConnections            string = "connections"
	PropNameCPULevel               string = "cpulevel"
	PropNameDateTime               string = "dateTime"    //
	PropNameDescription            string = "description" // Device description
	PropNameDeviceType             string = "deviceType"  // Device type from list below
	PropNameDewpoint               string = "dewpoint"
	PropNameDimmer                 string = "dimmer"
	PropNameDisabled               string = "disabled" // device or sensor is disabled
	PropNameDoorWindowSensor       string = "doorwindowsensor"
	PropNameElectricCurrent        string = "current"
	PropNameElectricEnergy         string = "energy"
	PropNameElectricPower          string = "power"
	PropNameErrors                 string = "errors"
	PropNameEvent                  string = "event" // Enable/disable event publishing
	//
	PropNameFilename       string = "filename"       // [string] filename to write images or other values to
	PropNameGatewayAddress string = "gatewayAddress" // [string] the 3rd party gateway address
	PropNameHostname       string = "hostname"       // [string] network device hostname
	PropNameHeatIndex      string = "heatindex"      // [number] unit=C or F
	PropNameHue            string = "hue"            //
	PropNameHumidex        string = "humidex"        // [number] unit=C or F
	PropNameHumidity       string = "humidity"       // [string] %
	PropNameImage          string = "image"          // [byteArray] unit=jpg,gif,png
	PropNameLatency        string = "latency"        // [number] sec, msec
	PropNameLatitude       string = "latitude"       // [number]
	PropNameLatLon         string = "latlon"         // [number,number] latitude, longitude pair of the device for display on a map r/w
	PropNameLevel          string = "level"          // [number] generic sensor level
	PropNameLongitude      string = "longitude"      // [number]
	PropNameLocalIP        string = "localIP"        // [string] for IP nodes
	PropNameLocation       string = "location"       // [string]
	PropNameLocationName   string = "locationName"   // [string] name of a location
	PropNameLock           string = "lock"           //
	PropNameLoginName      string = "loginName"      // [string] login name to connect to the device. Value is not published
	PropNameLuminance      string = "luminance"      // [number]
	PropNameMAC            string = "mac"            // [string] MAC address for IP nodes
	PropNameManufacturer   string = "manufacturer"   // [string] device manufacturer
	PropNameMax            string = "max"            // [number] maximum value of sensor or config
	PropNameMin            string = "min"            // [number] minimum value of sensor or config
	PropNameModel          string = "model"          // [string] device model
	PropNameMotion         string = "motion"         // [boolean]
	PropNameMute           string = "avmute"         // [boolean]
	PropNameName           string = "name"           // [string] Name of device or service
	PropNameNetmask        string = "netmask"        // [string] IP network mask
	PropNameOnOffSwitch    string = "switch"         // [boolean]
	//
	PropNamePassword        string = "password" // password to connect. Value is not published.
	PropNamePlay            string = "avplay"
	PropNamePollInterval    string = "pollInterval" // polling interval in seconds
	PropNamePort            string = "port"         // network address port
	PropNamePowerSource     string = "powerSource"  // battery, usb, mains
	PropNameProduct         string = "product"      // device product or model name
	PropNamePublicKey       string = "publicKey"    // public key for encrypting sensitive configuration settings
	PropNamePushButton      string = "pushbutton"   // with nr of pushes
	PropNameRain            string = "rain"
	PropNameRelay           string = "relay"
	PropNameSaturation      string = "saturation"
	PropNameScale           string = "scale"
	PropNameSignalStrength  string = "signalstrength"
	PropNameSmokeDetector   string = "smokedetector"
	PropNameSnow            string = "snow"
	PropNameSoftwareVersion string = "softwareVersion" // version of the software running the node
	PropNameSoundDetector   string = "sounddetector"
	PropNameSubnet          string = "subnet" // IP subnets configuration
	PropNameSwitch          string = "switch" // on/off switch: "on" "off"
	PropNameTemperature     string = "temperature"
	// PropNameType              string = "type" // Node type
	PropNameUltraviolet       string = "ultraviolet"
	PropNameUnknown           string = ""    // Not a known output
	PropNameURL               string = "url" // node URL
	PropNameVibrationDetector string = "vibrationdetector"
	PropNameValue             string = "value" // generic value
	PropNameVoltage           string = "voltage"
	PropNameVolume            string = "volume"
	PropNameWaterLevel        string = "waterlevel"
	PropNameWeather           string = "weather" // description of weather, eg sunny
	PropNameWindHeading       string = "windheading"
	PropNameWindSpeed         string = "windspeed"
)

// Standard ISO8601 timeformat
const TimeFormat = "2006-01-02T15:04:05.000-0700"

//TODO: Match with UN/CEFACT unitname codes as defined in:
//   https://www.unece.org/cefact.html
//   location codes: https://www.unece.org/cefact/locode/service/location.html
//
// UnitNameXyz defines constants with input and output unitname names.
const (
	UnitNameNone            string = ""
	UnitNameAmp             string = "A"
	UnitNameCelcius         string = "C"
	UnitNameCandela         string = "cd"
	UnitNameCount           string = "#"
	UnitNameDegree          string = "Degree"
	UnitNameFahrenheit      string = "F"
	UnitNameFeet            string = "ft"
	UnitNameGallon          string = "Gal"
	UnitNameJpeg            string = "jpeg"
	UnitNameKelvin          string = "K"
	UnitNameKmPerHour       string = "Kph"
	UnitNameLiter           string = "L"
	UnitNameMercury         string = "hg"
	UnitNameMeter           string = "m"
	UnitNameMetersPerSecond string = "m/s"
	UnitNameMilesPerHour    string = "mph"
	UnitNameMillibar        string = "mbar"
	UnitNameMole            string = "mol"
	UnitNamePartsPerMillion string = "ppm"
	UnitNamePng             string = "png"
	UnitNameKWH             string = "KWh"
	UnitNameKG              string = "kg"
	UnitNameLux             string = "lux"
	UnitNamePascal          string = "Pa"
	UnitNamePercent         string = "%"
	UnitNamePounds          string = "lbs"
	UnitNameSpeed           string = "m/s"
	UnitNamePSI             string = "psi"
	UnitNameSecond          string = "s"
	UnitNameVolt            string = "V"
	UnitNameWatt            string = "W"
)

// var (
// 	// UnitNameValuesAtmosphericPressure unitname values for atmospheric pressure
// 	UnitNameValuesAtmosphericPressure = fmt.Sprintf("%s, %s, %s", UnitNameMillibar, UnitNameMercury, UnitNamePSI)
// 	UnitNameValuesImage               = fmt.Sprintf("%s, %s", UnitNameJpeg, UnitNamePng)
// 	UnitNameValuesLength              = fmt.Sprintf("%s, %s", UnitNameMeter, UnitNameFeet)
// 	UnitNameValuesSpeed               = fmt.Sprintf("%s, %s, %s", UnitNameMetersPerSecond, UnitNameKmPerHour, UnitNameMilesPerHour)
// 	UnitNameValuesTemperature         = fmt.Sprintf("%s, %s", UnitNameCelcius, UnitNameFahrenheit)
// 	UnitNameValuesWeight              = fmt.Sprintf("%s, %s", UnitNameKG, UnitNamePounds)
// 	UnitNameValuesVolume              = fmt.Sprintf("%s, %s", UnitNameLiter, UnitNameGallon)
// )
