package metrics

import (
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var enabled atomic.Bool

var (
	scanned = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_scanned_total",
		Help: "Total number of scanned resources.",
	})
	violations = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_violations_total",
		Help: "Total number of violations.",
	})
	errors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_errors_total",
		Help: "Total number of errors.",
	})
	remediationAttempted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_remediation_attempted_total",
		Help: "Total number of remediation attempts.",
	})
	remediated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_remediated_total",
		Help: "Total number of successful remediations.",
	})
	remediationSkipped = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_remediation_skipped_total",
		Help: "Total number of skipped remediations.",
	})
	discoveryErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_discovery_errors_total",
		Help: "Total number of discovery errors.",
	})
	clientErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_client_errors_total",
		Help: "Total number of OpenStack client creation errors.",
	})
	serviceNotFound = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_service_not_found_total",
		Help: "Total number of service lookup failures.",
	})
	discovererNotFound = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_discoverer_not_found_total",
		Help: "Total number of discoverer lookup failures.",
	})
	auditorNotFound = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ospa_auditor_not_found_total",
		Help: "Total number of auditor lookup failures.",
	})
)

func init() {
	prometheus.MustRegister(
		scanned,
		violations,
		errors,
		remediationAttempted,
		remediated,
		remediationSkipped,
		discoveryErrors,
		clientErrors,
		serviceNotFound,
		discovererNotFound,
		auditorNotFound,
	)
}

func Enable() {
	enabled.Store(true)
}

func IncScanned() {
	if enabled.Load() {
		scanned.Inc()
	}
}

func IncViolations() {
	if enabled.Load() {
		violations.Inc()
	}
}

func IncErrors() {
	if enabled.Load() {
		errors.Inc()
	}
}

func IncRemediationAttempted() {
	if enabled.Load() {
		remediationAttempted.Inc()
	}
}

func IncRemediated() {
	if enabled.Load() {
		remediated.Inc()
	}
}

func IncRemediationSkipped() {
	if enabled.Load() {
		remediationSkipped.Inc()
	}
}

func IncDiscoveryErrors() {
	if enabled.Load() {
		discoveryErrors.Inc()
	}
}

func IncClientErrors() {
	if enabled.Load() {
		clientErrors.Inc()
	}
}

func IncServiceNotFound() {
	if enabled.Load() {
		serviceNotFound.Inc()
	}
}

func IncDiscovererNotFound() {
	if enabled.Load() {
		discovererNotFound.Inc()
	}
}

func IncAuditorNotFound() {
	if enabled.Load() {
		auditorNotFound.Inc()
	}
}

// StartServer starts the Prometheus metrics endpoint.
func StartServer(addr string) error {
	Enable()
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, mux)
}

