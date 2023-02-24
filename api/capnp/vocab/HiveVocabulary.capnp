# Package vocab with HiveOT iotschema vocabulary for sensor, actuator and unitname names
# TODO: base this of a universally accepted ontology. Closest is iotschema.org but that seems incomplete
@0xdb74211d971bf76f;

using Go = import "/go.capnp";
$Go.package("vocab");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

# Standard

const iSO8601Format :Text = "2006-01-02T15:04:05.999-0700";
# ISO8601Format standardized time format with msec resolution for use by Things using ISO8601 UTC


#const timeFormat :Text = "2006-01-02T15:04:05.000-0700";
# TimeFormat using standard ISO8601

# Vocabulary of device type names
const deviceTypeAlarm          :Text = "alarm";          # an alarm emitter
const deviceTypeAVControl      :Text = "avControl";      # Audio/Video controller
const deviceTypeAVReceiver     :Text = "avReceiver";     # Node is a (not so) smart radio/receiver/amp (eg, denon)
const deviceTypeBeacon         :Text = "beacon";         # device is a location beacon
const deviceTypeBinding        :Text = "binding";         # device is a protocol binding service
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

# Vocabulary of Standardized thing property names
const vocabAcceleration           :Text = "acceleration";
const vocabActive                 :Text = "active";
const vocabAddress                :Text = "address";     # device domain or ip address
const vocabAirQuality             :Text = "airQuality";  #
const vocabAlarm                  :Text = "alarm";       # state is alarm vs inactive
const vocabAlarmState             :Text = "alarmState";  #
const vocabAlarmType              :Text = "alarmType";   #
const vocabAtmosphericPressure    :Text = "atmosphericPressure"; #
const vocabBatch                  :Text = "batch";        # Batch publishing size
const vocabBatteryLevel           :Text = "batteryLevel"; # level % of battery
const vocabBatteryLow             :Text = "batteryLow";   # status of battery
const vocabCarbonDioxideLevel     :Text = "co2Level";     #
const vocabCarbonMonoxideDetector :Text = "coDetector";   #
const vocabCarbonMonoxideLevel    :Text = "coLevel";      #
const vocabChannel                :Text = "avChannel";    #
const vocabColor                  :Text = "color";        # Color in hex notation
const vocabColorTemperature       :Text = "colorTemperature"; #
const vocabConnections            :Text = "connections"; #
const vocabCPULevel               :Text = "cpuLevel";    #
const vocabDateTime               :Text = "dateTime";    #
const vocabDescription            :Text = "description"; # Device description
const vocabDeviceType             :Text = "deviceType";  # Device type from list below
const vocabDewpoint               :Text = "dewpoint";    #
const vocabDimmer                 :Text = "dimmer";      #
const vocabDuration               :Text = "duration";      #
const vocabDisabled               :Text = "disabled";    # device or sensor is disabled
const vocabDoorWindowSensor       :Text = "doorWindowSensor";  #
const vocabElectricCurrent        :Text = "current";     #
const vocabElectricEnergy         :Text = "energy";      #
const vocabElectricPower          :Text = "power";       #
const vocabErrors                 :Text = "errors";      #
const vocabEvent                  :Text = "event";       # Enable/disable event publishing
	
const vocabFilename        :Text = "filename";        # [string] filename to write images or other values to
const vocabFirmwareVersion :Text = "firmwareVersion"; # [string] firmware version name
const vocabGatewayAddress  :Text = "gatewayAddress";  # [string] the 3rd party gateway address
const vocabHardwareVersion :Text = "hardwareVersion"; # [string] the physical device version name
const vocabHostname       :Text = "hostname";        # [string] network device hostname
const vocabHeatIndex      :Text = "heatindex";       # [number] unit=C or F
const vocabHue            :Text = "hue";             #
const vocabHumidex        :Text = "humidex";       # [number] unit=C or F
const vocabHumidity       :Text = "humidity";      # [string] %
const vocabImage          :Text = "image";         # [byteArray] unit=jpg,gif,png
const vocabLatency        :Text = "latency";       # [number] sec, msec
const vocabLatitude       :Text = "latitude";      # [number]
const vocabLatLon         :Text = "latlon";        # [number,number] latitude, longitude pair of the device for display on a map r/w
const vocabLevel          :Text = "level";         # [number] generic sensor level
const vocabLongitude      :Text = "longitude";     # [number]
const vocabLocalIP        :Text = "localIP";       # [string] for IP nodes
const vocabLocation       :Text = "location";      # [string]
const vocabLocationName   :Text = "locationName";  # [string] name of a location
const vocabLock           :Text = "lock";          #
const vocabLoginName      :Text = "loginName";     # [string] login name to connect to the device. Value is not published
const vocabLuminance      :Text = "luminance";     # [number] in lumen
const vocabMAC            :Text = "mac";           # [string] MAC address for IP nodes
const vocabManufacturer   :Text = "manufacturer";  # [string] device manufacturer
const vocabMax            :Text = "max";           # [number] maximum value of sensor or config
const vocabMemory         :Text = "memory";        # [number] available/used/free memory
const vocabMin            :Text = "min";           # [number] minimum value of sensor or config
const vocabModel          :Text = "model";         # [string] device model
const vocabMotion         :Text = "motion";        # [boolean]
const vocabMute           :Text = "avMute";        # [boolean]
const vocabName           :Text = "name";          # [string] DisplayName of device or service
const vocabNetmask        :Text = "netmask";       # [string] IP network mask
const vocabOnOffSwitch    :Text = "switch";        # [boolean]
	
const vocabPassword          :Text = "password";    # password to connect. Value is not published.
const vocabPlay              :Text = "avPlay";      #
const vocabPollInterval      :Text = "pollInterval";# polling interval in seconds
const vocabPort              :Text = "port";        # network address port
const vocabPowerSource       :Text = "powerSource"; # battery, usb, mains
const vocabProduct           :Text = "product";     # device product or model name
const vocabPublicKey         :Text = "publicKey";   # public key for encrypting sensitive configuration settings
const vocabPushButton        :Text = "pushButton";  # with nr of pushes
const vocabRain              :Text = "rain";        #
const vocabRelay             :Text = "relay";       #
const vocabSaturation        :Text = "saturation";  #
const vocabScale             :Text = "scale";       #
const vocabSignalStrength    :Text = "signalStrength";  #
const vocabSmokeDetector     :Text = "smokeDetector";   #
const vocabSnow              :Text = "snow";            #
const vocabSoftwareVersion   :Text = "softwareVersion"; # version of the appplication software running the node
const vocabSoundDetector     :Text = "soundDetector";   #
const vocabSubnet            :Text = "subnet";          # IP subnets configuration
const vocabSwitch            :Text = "switch";          # on/off switch: "on";"off"
const vocabTemperature       :Text = "temperature";     #
const vocabTitle             :Text = "title";           # device title
const vocabUltraviolet       :Text = "ultraviolet";     #
const vocabUnknown           :Text = "";                # Not a known output
const vocabURL               :Text = "url";             # node URL
const vocabVibration         :Text = "vibration";       # vibration sensor
const vocabValue             :Text = "value";           # generic value
const vocabVoltage           :Text = "voltage";         # eg water flow in m3/sec or gpm
const vocabVolume            :Text = "volume";          #
const vocabWaterLevel        :Text = "waterLevel";      #
const vocabWeather           :Text = "weather";         # description of weather, eg sunny
const vocabWindHeading       :Text = "windHeading";      #
const vocabWindSpeed         :Text = "windSpeed";       #



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
