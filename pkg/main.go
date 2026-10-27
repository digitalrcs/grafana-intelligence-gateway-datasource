package main

import (
	"os"

	"github.com/DigitalRCS/grafana-intelligence-gateway-datasource/pkg/plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func manageOptions() datasource.ManageOpts {
	return datasource.ManageOpts{GRPCSettings: backend.GRPCSettings{
		// The SDK defaults to 16 MiB. Allow the gateway's 18 MiB image JSON
		// request plus bounded protobuf/data-source metadata headroom; the
		// resource handler still enforces its stricter application limits.
		MaxReceiveMsgSize: 20 << 20,
	}}
}

func main() {
	// Start listening to requests sent from Grafana. This call is blocking so
	// it won't finish until Grafana shuts down the process or the plugin choose
	// to exit by itself using os.Exit. Manage automatically manages life cycle
	// of datasource instances. It accepts datasource instance factory as first
	// argument. This factory will be automatically called on incoming request
	// from Grafana to create different instances of SampleDatasource (per datasource
	// ID). When datasource configuration changed Dispose method will be called and
	// new datasource instance created using NewSampleDatasource factory.
	if err := datasource.Manage("digitalrcs-intelligencegateway-datasource", plugin.NewDatasource, manageOptions()); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
