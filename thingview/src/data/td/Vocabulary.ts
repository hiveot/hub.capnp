// Definition of Thing TD fields 
// This javascript definition must be kept in sync with the golang hub/lib/client/pkg/vocab definitions

// Wost device types
export const DeviceTypeAlarm = "alarm"          // an alarm emitter
export const DeviceTypeAVControl = "avControl"      // Audio/Video controller
export const DeviceTypeAVReceiver = "avReceiver"     // Node is a (not so) smart radio/receiver/amp (eg, denon)
export const DeviceTypeBeacon = "beacon"         // device is a location beacon
export const DeviceTypeButton = "button"         // device is a physical button device with one or more buttons
export const DeviceTypeAdapter = "adapter"        // software adapter or service, eg virtual device
export const DeviceTypePhone = "phone"          // device is a phone
export const DeviceTypeCamera = "camera"         // Node with camera
export const DeviceTypeComputer = "computer"       // General purpose computer
export const DeviceTypeDimmer = "dimmer"         // light dimmer
export const DeviceTypeGateway = "gateway"        // Node is a gateway for other nodes (onewire, zwave, etc)
export const DeviceTypeKeypad = "keypad"         // Entry key pad
export const DeviceTypeLock = "lock"           // Electronic door lock
export const DeviceTypeMultisensor = "multisensor"    // Node with multiple sensors
export const DeviceTypeNetRepeater = "netRepeater"    // Node is a zwave or other network repeater
export const DeviceTypeNetRouter = "netRouter"      // Node is a network router
export const DeviceTypeNetSwitch = "netSwitch"      // Node is a network switch
export const DeviceTypeNetWifiAP = "wifiAP"         // Node is a wifi access point
export const DeviceTypeOnOffSwitch = "onOffSwitch"    // Node is a physical on/off switch
export const DeviceTypePowerMeter = "powerMeter"     // Node is a power meter
export const DeviceTypeSensor = "sensor"         // Node is a single sensor (volt,...)
export const DeviceTypeService = "service"        // Node provides a service
export const DeviceTypeSmartlight = "smartlight"     // Node is a smart light, eg philips hue
export const DeviceTypeThermometer = "thermometer"    // Node is a temperature meter
export const DeviceTypeThermostat = "thermostat"     // Node is a thermostat control unit
export const DeviceTypeTV = "tv"             // Node is a (not so) smart TV
export const DeviceTypeUnknown = "unknown"        // type not identified
export const DeviceTypeWallpaper = "wallpaper"      // Node is a wallpaper montage of multiple images
export const DeviceTypeWaterValve = "waterValve"     // Water valve control unit
export const DeviceTypeWeatherService = "weatherService" // Node is a service providing current and forecasted weather
export const DeviceTypeWeatherStation = "weatherStation" // Node is a weatherstation device
export const DeviceTypeWeighScale = "weighScale"     // Node is an electronic weight scale

// PropertyTypeActuator is an actuator (readonly, use Actions)
export const PropertyTypeActuator = "wost:actuator"
// PropertyTypeAttr is a readonly internal Thing attribute
export const PropertyTypeAttr = "wost:attr"
// PropertyTypeConfig is a writable configuration
export const PropertyTypeConfig = "wost:configuration"
// PropertyTypeSensor is a readonly sensor
export const PropertyTypeSensor = "wost:sensor"
// PropertyTypeState is a readonly internal state
export const PropertyTypeState = "wost:state"
// PropertyTypeInput is an input (use in Actions)
export const PropertyTypeInput = "wost:input"
// PropertyTypeOutput is an output (when different from sensor)
export const PropertyTypeOutput = "wost:output"

/**
 * WoST Vocabulary property names to be used by Things and plugins to build their TD
 */
export const PropNameAcceleration = "acceleration"
export const PropNameAddress = "address" // device domain or ip address
export const PropNameAirQuality = "airQuality"
export const PropNameAlarm = "alarm"
export const PropNameAtmosphericPressure = "atmosphericPressure"
export const PropNameBatch = "batch" // Batch publishing size
export const PropNameBattery = "battery"
export const PropNameCarbonDioxideLevel = "co2level"
export const PropNameCarbonMonoxideDetector = "coDetector"
export const PropNameCarbonMonoxideLevel = "coLevel"
export const PropNameChannel = "avChannel"
export const PropNameColor = "color" // Color in hex notation
export const PropNameColorTemperature = "colortemperature"
export const PropNameConnections = "connections"
export const PropNameCPULevel = "cpuLevel"
export const PropNameDateTime = "dateTime"    //
export const PropNameDescription = "description" // Device description
export const PropNameDeviceType = "deviceType"  // Device type from list below
export const PropNameDewpoint = "dewpoint"
export const PropNameDimmer = "dimmer"
export const PropNameDisabled = "disabled" // device or sensor is disabled
export const PropNameDoorWindowSensor = "doorWindowSensor"
export const PropNameElectricCurrent = "current"
export const PropNameElectricEnergy = "energy"
export const PropNameElectricPower = "power"
export const PropNameErrors = "errors"
export const PropNameEvent = "event" // Enable/disable event publishing
//
export const PropNameFilename = "filename"       // [string] filename to write images or other values to
export const PropNameGatewayAddress = "gatewayAddress" // [string] the 3rd party gateway address
export const PropNameHostname = "hostname"       // [string] network device hostname
export const PropNameHeatIndex = "heatindex"      // [number] unit=C or F
export const PropNameHue = "hue"            //
export const PropNameHumidex = "humidex"        // [number] unit=C or F
export const PropNameHumidity = "humidity"       // [string] %
export const PropNameImage = "image"          // [byteArray] unit=jpg,gif,png
export const PropNameLatency = "latency"        // [number] sec, msec
export const PropNameLatitude = "latitude"       // [number]
export const PropNameLatLon = "latlon"         // [number,number] latitude, longitude pair of the device for display on a map r/w
export const PropNameLevel = "level"          // [number] generic sensor level
export const PropNameLongitude = "longitude"      // [number]
export const PropNameLocalIP = "localIP"        // [string] for IP nodes
export const PropNameLocation = "location"       // [string]
export const PropNameLocationName = "locationName"   // [string] name of a location
export const PropNameLock = "lock"           //
export const PropNameLoginName = "loginName"      // [string] login name to connect to the device. Value is not published
export const PropNameLuminance = "luminance"      // [number]
export const PropNameMAC = "mac"            // [string] MAC address for IP nodes
export const PropNameManufacturer = "manufacturer"   // [string] device manufacturer
export const PropNameMax = "max"            // [number] maximum value of sensor or config
export const PropNameMin = "min"            // [number] minimum value of sensor or config
export const PropNameModel = "model"          // [string] device model
export const PropNameMotion = "motion"         // [boolean]
export const PropNameMute = "avMute"         // [boolean]
export const PropNameName = "name"           // [string] Name of device or service
export const PropNameNetmask = "netmask"        // [string] IP network mask
export const PropNameOnOffSwitch = "switch"         // [boolean]
//
export const PropNamePassword = "password" // password to connect. Value is not published.
export const PropNamePlay = "avPlay"
export const PropNamePollInterval = "pollInterval" // polling interval in seconds
export const PropNamePort = "port"         // network address port
export const PropNamePowerSource = "powerSource"  // battery, usb, mains
export const PropNameProduct = "product"      // device product or model name
export const PropNamePublicKey = "publicKey"    // public key for encrypting sensitive configuration settings
export const PropNamePushButton = "pushButton"   // with nr of pushes
export const PropNameRain = "rain"
export const PropNameRelay = "relay"
export const PropNameSaturation = "saturation"
export const PropNameScale = "scale"
export const PropNameSignalStrength = "signalStrength"
export const PropNameSmokeDetector = "smokeDetector"
export const PropNameSnow = "snow"
export const PropNameSoftwareVersion = "softwareVersion" // version of the software running the node
export const PropNameSoundDetector = "soundDetector"
export const PropNameSubnet = "subnet" // IP subnets configuration
export const PropNameSwitch = "switch" // on/off switch: "on" "off"
export const PropNameTemperature = "temperature"
// export const PropNameType              = "type" // Node type
export const PropNameUltraviolet = "ultraviolet"
export const PropNameUnknown = ""    // Not a known output
export const PropNameURL = "url" // node URL
export const PropNameVibrationDetector = "vibrationDetector"
export const PropNameValue = "value" // generic value
export const PropNameVoltage = "voltage"
export const PropNameVolume = "volume"
export const PropNameWaterLevel = "waterLevel"
export const PropNameWeather = "weather" // description of weather, eg sunny
export const PropNameWindHeading = "windHeading"
export const PropNameWindSpeed = "windSpeed"


/**
 * The following terms are defined in the WoT Thing Description definition
 */
// WoT data schema
export const WoTAtType = "@type"
export const WoTAtContext = "@context"
export const WoTAnyURI = "https://www.w3.org/2019/wot/td/v1"
export const WoTActions = "actions"
export const WoTCreated = "created"
export const WoTDescription = "description"
export const WoTDescriptions = "descriptions"
export const WoTEvents = "events"
export const WoTForms = "forms"
export const WoTID = "id"
export const WoTLinks = "links"
export const WoTProperties = "properties"
export const WoTSecurity = "security"
export const WoTSupport = "support"
export const WoTTitle = "title"
export const WoTTitles = "titles"
export const WoTVersion = "version"

// additional data schema vocab
export const WoTConst = "const"
export const WoTDataTypeAnyURI = "anyURI" // simple type
export const WoTDataTypeArray = "array"
export const WoTDataTypeBool = "boolean"  // simple type
export const WoTDataTypeDateTime = "dateTime" // ISO8601: YYYY-MM-DDTHH:MM:SS.sss[-TZ|+TZ|z]

export const WoTDataTypeInteger = "integer"     // simple type
export const WoTDataTypeUnsignedInt = "unsignedInt" // simple type
export const WoTDataTypeNumber = "number"
export const WoTDataTypeObject = "object"
export const WoTDataType = "string" // simple type
// WoTDouble              = "double" // min, max of number are doubles
export const WoTEnum = "enum"
export const WoTFormat = "format"
export const WoTHref = "href"
export const WoTInput = "input"
export const WoTMaximum = "maximum"
export const WoTMaxItems = "maxItems"
export const WoTMaxLength = "maxLength"
export const WoTMinimum = "minimum"
export const WoTMinItems = "minItems"
export const WoTMinLength = "minLength"
export const WoTModified = "modified"
export const WoTOperation = "op"
export const WoTOutput = "output"
export const WoTReadOnly = "readOnly"
export const WoTRequired = "required"
export const WoTWriteOnly = "writeOnly"
export const WoTUnit = "unit"
