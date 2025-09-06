package gpmiddleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var defaultMetricPath = "/metrics"

// RequestCounterURLLabelMappingFn url label
type RequestCounterURLLabelMappingFn func(c *gin.Context) string

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqDur        *prometheus.HistogramVec
	router        *gin.Engine
	listenAddress string
	MetricsPath   string
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(subsystem string) *Prometheus {
	p := &Prometheus{
		MetricsPath: defaultMetricPath,
	}

	p.registerMetrics(subsystem)

	return p
}

// SetListenAddress for exposing metrics on address. If not set, it will be exposed at the
// same address of the gin engine that is being used
func (p *Prometheus) SetListenAddress(address string) {
	p.listenAddress = address
	if p.listenAddress != "" {
		p.router = gin.Default()
	}
}

// SetListenAddressWithRouter for using a separate router to expose metrics. (this keeps things like GET /metrics out of
// your content's access log).
func (p *Prometheus) SetListenAddressWithRouter(listenAddress string, r *gin.Engine) {
	p.listenAddress = listenAddress
	if len(p.listenAddress) > 0 {
		p.router = r
	}
}

// SetMetricsPath set metrics paths
func (p *Prometheus) SetMetricsPath(e *gin.Engine) {
	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, prometheusHandler())
	}
}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go p.router.Run(p.listenAddress)
	}
}

func (p *Prometheus) registerMetrics(subsystem string) {
	p.reqDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "Histogram request latencies",
			Buckets:   []float64{0.1, 0.2, 0.3, 0.5, 0.75, 1, 1.5, 2, 3, 5},
		},
		[]string{"code", "path"},
	)

	prometheus.Register(p.reqDur)
}

// HandlerFunc defines handler function for middleware
func (p *Prometheus) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.String() == p.MetricsPath {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)

		path := c.FullPath()
		if path == "" { // path empty -> no route found
			path = "404"
		}
		p.reqDur.WithLabelValues(status, c.Request.Method+"_"+path).Observe(elapsed)
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Use adds the middleware to a gin engine with /metrics route path.
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc())
	e.GET(p.MetricsPath, prometheusHandler())
}

// UseCustom adds the middleware to a gin engine with a custom route path.
func (p *Prometheus) UseCustom(e *gin.Engine) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPath(e)
}
