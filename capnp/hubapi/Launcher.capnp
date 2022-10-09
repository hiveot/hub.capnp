# Cap'n proto definition for the service launcher
@0xe42f87955bd521e9;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub.capnp/go/hubapi");

struct ServiceInfo {
	cpu @0 :Int32;
	# CPU usage in %. 0 when not running

	rss @1 :Int64;
	# RSS (Real) Memory usage in bytes. 0 when not running

	error @2 :Text;
	# The last error status when running

	modifiedTime @3 :Text;
	# Service modified time ISO8601

	name @4 :Text;
	# name of the service

	path @5 :Text;
	# path to service executable

	pid @6 :Int32;
	# Program PID when started. This remains after stopping.

	startCount @7 :Int32;
	# Number of times service was (re)started

	startTime @8 :Text;
	# Starting time of the service in ISO8601

	stopTime @9 :Text;
	# Stopped time of the service in ISO8601

	running @10 :Bool;
	# Service is currently running

	size @11 :Int64;
	# binary size of the service in bytes

	uptime @12 :Int32;
	# uptime time the service is running in seconds.
}

interface CapLauncher {
  # Service launching capabilities

  list @0 () -> (infoList :List(ServiceInfo));
  # List the available services and their status

  start @1 (name :Text) -> (info :ServiceInfo);
  # Start the service with the given name. The service must exist in the result of List.

  stop @2 (name :Text) -> (info :ServiceInfo);
  # Stop a service that was previously started.

  stopAll @3 () -> ();
  # Stop all running services
}
