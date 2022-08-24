package telemetry

import (
	"os"

	"github.com/pyroscope-io/client/pyroscope"
	"github.com/yolkhovyy/golang-grpc-demo/config"
)

func PyroscopeStart(config config.ProfilerConfig) (*pyroscope.Profiler, error) {
	if !config.Enabled {
		return nil, nil
	}
	config.Pyroscope.AuthToken = os.Getenv("PYROSCOPE_AUTH_TOKEN")
	// config.Pyroscope.Logger = pyroscope.StandardLogger
	config.Pyroscope.Logger = nil
	return pyroscope.Start(config.Pyroscope)
}
