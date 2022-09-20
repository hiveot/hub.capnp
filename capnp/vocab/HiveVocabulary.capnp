# Package vocab with HiveOT iotschema vocabulary for sensor, actuator and unitname names
# TODO: base this of a universally accepted ontology. Closest is iotschema.org but that seems incomplete
@0xdb74211d971bf76f;

using Go = import "/go.capnp";
$Go.package("vocab");
$Go.import("github.com/hiveot/hub.capnp/go/vocab");

# Standard

const timeFormat :Text = "2006-01-02T15:04:05.000-0700";
# TimeFormat using standard ISO8601



# Vocabulary of device type names
const deviceTypeAlarm          :Text = "alarm";          # an alarm emitter
const deviceTypeAVControl      :Text = "avControl";      # Audio/Video controller
const deviceTypeAVReceiver     :Text = "avReceiver";     # Node is a (not so) smart radio/receiver/amp (eg, denon)
const deviceTypeBeacon         :Text = "beacon";         # device is a location beacon
const deviceTypeButton         :Text = "button";         # device is a physical button device with one or more buttons
const deviceTypeAdapter        :Text = "adapter";        # software adapter or service, eg virtual device
const deviceTypePhone          :Text = "phone";          # device is a phone
const deviceTypeCamera         :Text = "camera";         # Node with camera
const deviceTypeComputer       :Text = "computer";       # General purpose computer
const deviceTypeDimmer         :Text = "dimmer";         # light dimmer
const deviceTypeGateway        :Text = "gateway";        # Node is a gateway for other nodes (onewire, zwave, etc)
const deviceTypeKeypad         :Text = "keypad";         # Entry key pad
const deviceTypeLock           :Text = "lock";           # Electronic door lock
const deviceTypeMultisensor    :Text = "multisensor";    # Node with multiple sensors
const deviceTypeNetRepeater    :Text = "netRepeater";    # Node is a zwave or other network repeater
const deviceTypeNetRouter      :Text = "netRouter";      # Node is a network router
const deviceTypeNetSwitch      :Text = "netSwitch";      # Node is a network switch
const deviceTypeNetWifiAP      :Text = "wifiAP";         # Node is a wifi access point
const deviceTypeOnOffSwitch    :Text = "onOffSwitch";    # Node is a physical on/off switch
const deviceTypePowerMeter     :Text = "powerMeter";     # Node is a power meter
const deviceTypeSensor         :Text = "sensor";         # Node is a single sensor (volt,...)
const deviceTypeService        :Text = "service";        # Node provides a service
const deviceTypeSmartlight     :Text = "smartlight";     # Node is a smart light, eg philips hue
const deviceTypeThermometer    :Text = "thermometer";    # Node is a temperature meter
const deviceTypeThermostat     :Text = "thermostat";     # Node is a thermostat control unit
const deviceTypeTV             :Text = "tv";             # Node is a (not so) smart TV
const deviceTypeUnknown        :Text = "unknown";        # type not identified
const deviceTypeWallpaper      :Text = "wallpaper";      # Node is a wallpaper montage of multiple images
const deviceTypeWaterValve     :Text = "waterValve";     # Water valve control unit
const deviceTypeWeatherService :Text = "weatherService"; # Node is a service providing current and forecasted weather
const deviceTypeWeatherStation :Text = "weatherStation"; # Node is a weatherstation device
const deviceTypeWeighScale     :Text = "weighScale";     # Node is an electronic weight scale

# Vocabulary of Thing property names 
const propNameAcceleration           :Text = "acceleration";
const propNameAddress                :Text = "address";     # device domain or ip address
const propNameAirQuality             :Text = "airQuality";  #
const propNameAlarm                  :Text = "alarm";       #
const propNameAtmosphericPressure    :Text = "atmosphericPressure"; #
const propNameBatch                  :Text = "batch";       # Batch publishing size
const propNameBattery                :Text = "battery";     #
const propNameCarbonDioxideLevel     :Text = "co2level";    #
const propNameCarbonMonoxideDetector :Text = "coDetector";  #
const propNameCarbonMonoxideLevel    :Text = "coLevel";     #
const propNameChannel                :Text = "avChannel";   #
const propNameColor                  :Text = "color";       # Color in hex notation
const propNameColorTemperature       :Text = "colorTemperature"; #
const propNameConnections            :Text = "connections"; #
const propNameCPULevel               :Text = "cpuLevel";    #
const propNameDateTime               :Text = "dateTime";    #
const propNameDescription            :Text = "description"; # Device description
const propNameDeviceType             :Text = "deviceType";  # Device type from list below
const propNameDewpoint               :Text = "dewpoint";    #
const propNameDimmer                 :Text = "dimmer";      #
const propNameDisabled               :Text = "disabled";    # device or sensor is disabled
const propNameDoorWindowSensor       :Text = "doorWindowSensor";  #
const propNameElectricCurrent        :Text = "current";     #
const propNameElectricEnergy         :Text = "energy";      #
const propNameElectricPower          :Text = "power";       #
const propNameErrors                 :Text = "errors";      #
const propNameEvent                  :Text = "event";       # Enable/disable event publishing
	
const propNameFilename       :Text = "filename";      # [string] filename to write images or other values to
const propNameGatewayAddress :Text = "gatewayAddress";# [string] the 3rd party gateway address
const propNameHostname       :Text = "hostname";      # [string] network device hostname
const propNameHeatIndex      :Text = "heatindex";     # [number] unit=C or F
const propNameHue            :Text = "hue";           #
const propNameHumidex        :Text = "humidex";       # [number] unit=C or F
const propNameHumidity       :Text = "humidity";      # [string] %
const propNameImage          :Text = "image";         # [byteArray] unit=jpg,gif,png
const propNameLatency        :Text = "latency";       # [number] sec, msec
const propNameLatitude       :Text = "latitude";      # [number]
const propNameLatLon         :Text = "latlon";        # [number,number] latitude, longitude pair of the device for display on a map r/w
const propNameLevel          :Text = "level";         # [number] generic sensor level
const propNameLongitude      :Text = "longitude";     # [number]
const propNameLocalIP        :Text = "localIP";       # [string] for IP nodes
const propNameLocation       :Text = "location";      # [string]
const propNameLocationName   :Text = "locationName";  # [string] name of a location
const propNameLock           :Text = "lock";          #
const propNameLoginName      :Text = "loginName";     # [string] login name to connect to the device. Value is not published
const propNameLuminance      :Text = "luminance";     # [number]
const propNameMAC            :Text = "mac";           # [string] MAC address for IP nodes
const propNameManufacturer   :Text = "manufacturer";  # [string] device manufacturer
const propNameMax            :Text = "max";           # [number] maximum value of sensor or config
const propNameMin            :Text = "min";           # [number] minimum value of sensor or config
const propNameModel          :Text = "model";         # [string] device model
const propNameMotion         :Text = "motion";        # [boolean]
const propNameMute           :Text = "avMute";        # [boolean]
const propNameName           :Text = "name";          # [string] DisplayName of device or service
const propNameNetmask        :Text = "netmask";       # [string] IP network mask
const propNameOnOffSwitch    :Text = "switch";        # [boolean]
	
const propNamePassword          :Text = "password";    # password to connect. Value is not published.
const propNamePlay              :Text = "avPlay";      #
const propNamePollInterval      :Text = "pollInterval";# polling interval in seconds
const propNamePort              :Text = "port";        # network address port
const propNamePowerSource       :Text = "powerSource"; # battery, usb, mains
const propNameProduct           :Text = "product";     # device product or model name
const propNamePublicKey         :Text = "publicKey";   # public key for encrypting sensitive configuration settings
const propNamePushButton        :Text = "pushButton";  # with nr of pushes
const propNameRain              :Text = "rain";        #
const propNameRelay             :Text = "relay";       #
const propNameSaturation        :Text = "saturation";  #
const propNameScale             :Text = "scale";       #
const propNameSignalStrength    :Text = "signalStrength";  #
const propNameSmokeDetector     :Text = "smokeDetector";   #
const propNameSnow              :Text = "snow";            #
const propNameSoftwareVersion   :Text = "softwareVersion"; # version of the software running the node
const propNameSoundDetector     :Text = "soundDetector";   #
const propNameSubnet            :Text = "subnet";          # IP subnets configuration
const propNameSwitch            :Text = "switch";          # on/off switch: "on";"off"
const propNameTemperature       :Text = "temperature";     #
const propNameTitle             :Text = "title";           # device title
const propNameUltraviolet       :Text = "ultraviolet";     #
const propNameUnknown           :Text = "";                # Not a known output
const propNameURL               :Text = "url";             # node URL
const propNameVibrationDetector :Text = "vibrationDetector"; #
const propNameValue             :Text = "value";           # generic value
const propNameVoltage           :Text = "voltage";         #
const propNameVolume            :Text = "volume";          #
const propNameWaterLevel        :Text = "waterLevel";      #
const propNameWeather           :Text = "weather";         # description of weather, eg sunny
const propNameWindHeading       :Text = "windHeading";      #
const propNameWindSpeed         :Text = "windSpeed";       #



# Unit constants
# TODO: Reconcile against UN/CEFACT unitname codes as defined in:
#   https://www.unece.org/cefact.html
#   location codes: https://www.unece.org/cefact/locode/service/location.html
#
# UnitNameXyz defines constants with input and output unitname names.
const unitNameNone            :Text = "";
const unitNameAmp             :Text = "A";
const unitNameCelcius         :Text = "C";
const unitNameCandela         :Text = "cd";
const unitNameCount           :Text = "#";
const unitNameDegree          :Text = "Degree";
const unitNameFahrenheit      :Text = "F";
const unitNameFeet            :Text = "ft";
const unitNameGallon          :Text = "Gal";
const unitNameJpeg            :Text = "jpeg";
const unitNameKelvin          :Text = "K";
const unitNameKmPerHour       :Text = "Kph";
const unitNameLiter           :Text = "L";
const unitNameMercury         :Text = "hg";
const unitNameMeter           :Text = "m";
const unitNameMetersPerSecond :Text = "m/s";
const unitNameMilesPerHour    :Text = "mph";
const unitNameMillibar        :Text = "mbar";
const unitNameMole            :Text = "mol";
const unitNamePartsPerMillion :Text = "ppm";
const unitNamePng             :Text = "png";
const unitNameKWH             :Text = "KWh";
const unitNameKG              :Text = "kg";
const unitNameLux             :Text = "lux";
const unitNamePascal          :Text = "Pa";
const unitNamePercent         :Text = "%";
const unitNamePounds          :Text = "lbs";
const unitNameSpeed           :Text = "m/s";
const unitNamePSI             :Text = "psi";
const unitNameSecond          :Text = "s";
const unitNameVolt            :Text = "V";
const unitNameWatt            :Text = "W";
