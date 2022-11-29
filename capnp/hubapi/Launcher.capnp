# Cap'n proto definition for the service launcher
@0xe42f87955bd521e9;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");
using Service = import "Service.capnp";

struct ServiceInfo {
	cpu @0 :Int32;
	# CPU usage in %. 0 when not running

	rss @1 :Int64;
	# RSS (Real) Memory usage in bytes. 0 when not running

	modifiedTime @2 :Text;
	# Service modified time ISO8601

	name @3 :Text;
	# name of the service

	path @4 :Text;
	# path to service executable

	pid @5 :Int32;
	# Program PID when started. This remains after stopping.

	running @6 :Bool;
	# Service is currently running

	size @7 :Int64;
	# binary size of the service in bytes

	status @8 :Text;
	# The last received status message

	startCount @9 :Int32;
	# Number of times service was restarted

	startTime @10 :Text;
	# Starting time of the service in ISO8601

	stopTime @11 :Text;
	# Stopped time of the service in ISO8601

	uptime @12 :Int32;
	# uptime time the service is running in seconds.
}

interface CapLauncher extends (Service.CapHiveOTService) {
  # Service launching capabilities

  list @0 (onlyRunning :Bool) -> (infoList :List(ServiceInfo));
  # List all available or only the running services and their status

  start @1 (name :Text) -> (info :ServiceInfo);
  # Start the service with the given name. The service must exist in the result of List.

  stop @2 (name :Text) -> (info :ServiceInfo);
  # Stop a service that was previously started.

  stopAll @3 () -> ();
  # Stop all running services
}
