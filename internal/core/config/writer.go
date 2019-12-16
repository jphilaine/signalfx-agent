package config

import (
	"net/url"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/signalfx/signalfx-agent/internal/core/dpfilters"
	"github.com/signalfx/signalfx-agent/internal/core/propfilters"
	log "github.com/sirupsen/logrus"
)

// WriterConfig holds configuration for the datapoint writer.
type WriterConfig struct {
	// The maximum number of datapoints to include in a batch before sending the
	// batch to the ingest server.  Smaller batch sizes than this will be sent
	// if datapoints originate in smaller chunks.
	DatapointMaxBatchSize int `yaml:"datapointMaxBatchSize" default:"1000"`
	// The maximum number of datapoints that are allowed to be buffered in the
	// agent (i.e. received from a monitor but have not yet received
	// confirmation of successful receipt by the target ingest/gateway server
	// downstream).  Any datapoints that come in beyond this number will
	// overwrite existing datapoints if they have not been sent yet, starting
	// with the oldest.
	MaxDatapointsBuffered int `yaml:"maxDatapointsBuffered" default:"25000"`
	// The analogue of `datapointMaxBatchSize` for trace spans.
	TraceSpanMaxBatchSize int `yaml:"traceSpanMaxBatchSize" default:"1000"`
	// Format to export traces in. Choices are "sfx" and "sapm"
	TraceExportFormat string `yaml:"traceExportFormat" default:"sfx"`
	// Deprecated: use `maxRequests` instead.
	DatapointMaxRequests int `yaml:"datapointMaxRequests"`
	// The maximum number of concurrent requests to make to a single ingest server
	// with datapoints/events/trace spans.  This number multiplied by
	// `datapointMaxBatchSize` is more or less the maximum number of datapoints
	// that can be "in-flight" at any given time.  Same thing for the
	// `traceSpanMaxBatchSize` option and trace spans.
	MaxRequests int `yaml:"maxRequests" default:"10"`
	// The agent does not send events immediately upon a monitor generating
	// them, but buffers them and sends them in batches.  The lower this
	// number, the less delay for events to appear in SignalFx.
	EventSendIntervalSeconds int `yaml:"eventSendIntervalSeconds" default:"1"`
	// The analogue of `maxRequests` for dimension property requests.
	PropertiesMaxRequests uint `yaml:"propertiesMaxRequests" default:"20"`
	// How many dimension property updates to hold pending being sent before
	// dropping subsequent property updates.  Property updates will be resent
	// eventually and they are slow to change so dropping them (esp on agent
	// start up) usually isn't a big deal.
	PropertiesMaxBuffered uint `yaml:"propertiesMaxBuffered" default:"10000"`
	// How long to wait for property updates to be sent once they are
	// generated.  Any duplicate updates to the same dimension within this time
	// frame will result in the latest property set being sent.  This helps
	// prevent spurious updates that get immediately overwritten by very flappy
	// property generation.
	PropertiesSendDelaySeconds uint `yaml:"propertiesSendDelaySeconds" default:"30"`
	// Properties that are synced to SignalFx are cached to prevent duplicate
	// requests from being sent, causing unnecessary load on our backend.
	PropertiesHistorySize uint `yaml:"propertiesHistorySize" default:"10000"`
	// If the log level is set to `debug` and this is true, all datapoints
	// generated by the agent will be logged.
	LogDatapoints bool `yaml:"logDatapoints"`
	// The analogue of `logDatapoints` for events.
	LogEvents bool `yaml:"logEvents"`
	// The analogue of `logDatapoints` for trace spans.
	LogTraceSpans bool `yaml:"logTraceSpans"`
	// If `true`, dimension updates will be logged at the INFO level.
	LogDimensionUpdates bool `yaml:"logDimensionUpdates"`
	// If true, and the log level is `debug`, filtered out datapoints will be
	// logged.
	LogDroppedDatapoints bool `yaml:"logDroppedDatapoints"`
	// Whether to send host correlation metrics to correlation traced services
	// with the underlying host
	SendTraceHostCorrelationMetrics *bool `yaml:"sendTraceHostCorrelationMetrics" default:"true"`
	// How long to wait after a trace span's service name is last seen to
	// continue sending the correlation datapoints for that service.  This
	// should be a duration string that is accepted by
	// https://golang.org/pkg/time/#ParseDuration.  This option is irrelvant if
	// `sendTraceHostCorrelationMetrics` is false.
	StaleServiceTimeout time.Duration `yaml:"staleServiceTimeout" default:"5m"`
	// How frequently to send host correlation metrics that are generated from
	// the service name seen in trace spans sent through or by the agent.  This
	// should be a duration string that is accepted by
	// https://golang.org/pkg/time/#ParseDuration.  This option is irrelvant if
	// `sendTraceHostCorrelationMetrics` is false.
	TraceHostCorrelationMetricsInterval time.Duration `yaml:"traceHostCorrelationMetricsInterval" default:"1m"`
	// How many trace spans are allowed to be in the process of sending.  While
	// this number is exceeded, the oldest spans will be discarded to
	// accommodate new spans generated to avoid memory exhaustion.  If you see
	// log messages about "Aborting pending trace requests..." or "Dropping new
	// trace spans..." it means that the downstream target for traces is not
	// able to accept them fast enough. Usually if the downstream is offline
	// you will get connection refused errors and most likely spans will not
	// build up in the agent (there is no retry mechanism). In the case of slow
	// downstreams, you might be able to increase `maxRequests` to increase the
	// concurrent stream of spans downstream (if the target can make efficient
	// use of additional connections) or, less likely, increase
	// `traceSpanMaxBatchSize` if your batches are maxing out (turn on debug
	// logging to see the batch sizes being sent) and being split up too much.
	// If neither of those options helps, your downstream is likely too slow to
	// handle the volume of trace spans and should be upgraded to more powerful
	// hardware/networking.
	MaxTraceSpansInFlight uint `yaml:"maxTraceSpansInFlight" default:"100000"`
	// The following are propagated from elsewhere
	HostIDDims          map[string]string      `yaml:"-"`
	IngestURL           string                 `yaml:"-"`
	APIURL              string                 `yaml:"-"`
	TraceEndpointURL    string                 `yaml:"-"`
	SignalFxAccessToken string                 `yaml:"-"`
	GlobalDimensions    map[string]string      `yaml:"-"`
	MetricsToInclude    []MetricFilter         `yaml:"-"`
	MetricsToExclude    []MetricFilter         `yaml:"-"`
	PropertiesToExclude []PropertyFilterConfig `yaml:"-"`
}

func (wc *WriterConfig) initialize() {
	if wc.DatapointMaxRequests != 0 {
		wc.MaxRequests = wc.DatapointMaxRequests
	} else {
		wc.DatapointMaxRequests = wc.MaxRequests
	}
}

// ParsedIngestURL parses and returns the ingest URL
func (wc *WriterConfig) ParsedIngestURL() *url.URL {
	ingestURL, err := url.Parse(wc.IngestURL)
	if err != nil {
		panic("IngestURL was supposed to be validated already")
	}
	return ingestURL
}

// ParsedAPIURL parses and returns the API server URL
func (wc *WriterConfig) ParsedAPIURL() *url.URL {
	apiURL, err := url.Parse(wc.APIURL)
	if err != nil {
		panic("apiUrl was supposed to be validated already")
	}
	return apiURL
}

// ParsedTraceEndpointURL parses and returns the trace endpoint server URL
func (wc *WriterConfig) ParsedTraceEndpointURL() *url.URL {
	if wc.TraceEndpointURL != "" {
		traceEndpointURL, err := url.Parse(wc.TraceEndpointURL)
		if err != nil {
			panic("traceEndpointUrl was supposed to be validated already")
		}
		return traceEndpointURL
	}
	return nil
}

// DatapointFilters creates the filter set for datapoints
func (wc *WriterConfig) DatapointFilters() (*dpfilters.FilterSet, error) {
	return makeOldFilterSet(wc.MetricsToExclude, wc.MetricsToInclude)
}

// PropertyFilters creates the filter set for dimension properties
func (wc *WriterConfig) PropertyFilters() (*propfilters.FilterSet, error) {
	return makePropertyFilterSet(wc.PropertiesToExclude)
}

// Hash calculates a unique hash value for this config struct
func (wc *WriterConfig) Hash() uint64 {
	hash, err := hashstructure.Hash(wc, nil)
	if err != nil {
		log.WithError(err).Error("Could not get hash of WriterConfig struct")
		return 0
	}
	return hash
}
