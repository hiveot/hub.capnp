package launcher

import "context"

// ServiceName used to connect to this service
const ServiceName = "launcher"

// ServiceInfo contains the running status of a service
type ServiceInfo struct {
	// CPU usage in %. 0 when not running
	CPU int

	// Memory usage in MB. 0 when not running
	MEM int

	// The last error status when running
	Error string

	// Service modified time ISO8601
	ModifiedTime string

	// name of the service
	Name string

	// path to service executable
	Path string

	// Program PID when started. This remains after stopping.
	PID int

	// Number of times the service was started (but might have stopped)
	StartCount int

	// Starting time of the service in ISO8601
	StartTime string

	// Stopped time of the service in ISO8601
	StopTime string

	// Service is currently running
	Running bool

	// binary size of the service in bytes
	Size int64

	// uptime time the service is running in seconds.
	Uptime int
}

// ILauncher defines the POGS based interface of the launcher service
type ILauncher interface {

	// List services
	List(ctx context.Context) ([]ServiceInfo, error)

	// Start a service
	Start(ctx context.Context, name string) (ServiceInfo, error)

	// Stop a running service
	Stop(ctx context.Context, name string) (ServiceInfo, error)

	// StopAll running services
	StopAll(ctx context.Context) error
}
