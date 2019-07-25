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
	reqCnt        *prometheus.CounterVec
	reqDur        *prometheus.HistogramVec
	router        *gin.Engine
	listenAddress string
	MetricsPath   string
	ReqCntURLLabelMappingFn RequestCounterURLLabelMappingFn
	// gin.Context string to use as a prometheus URL label
	URLLabelFromContext string
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(subsystem string) *Prometheus {

	p := &Prometheus{
		MetricsPath: defaultMetricPath,
		ReqCntURLLabelMappingFn: func(c *gin.Context) string {
			return c.Request.URL.String() // i.e. by default do nothing, i.e. return URL as is
		},
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
	
	p.reqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "request_count",
			Help:      "Number of request",
		},
		[]string{"code", "path", "handler", "host", "url"},
	)

	p.reqDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "request latencies",
			Buckets:   []float64{.0005, .001, .002, 0.004, .006, 0.008, .01, 0.015, .025, 0.04, .06, .08, 0.1, 0.15, 0.2, 0.3, 0.5},
		},
		[]string{"code", "path", "handler", "host", "url"},
	)

	prometheus.Register(p.reqCnt)
	prometheus.Register(p.reqDur)

}

// Use adds the middleware to a gin engine.
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPath(e)
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
		

		url := p.ReqCntURLLabelMappingFn(c)
		// sidecar specific mod
		if len(p.URLLabelFromContext) > 0 {
			u, found := c.Get(p.URLLabelFromContext)
			if !found {
				u = "unknown"
			}
			url = u.(string)
		}
		p.reqDur.WithLabelValues(status, c.Request.Method+"_"+c.Request.URL.Path, c.HandlerName(), c.Request.Host, url).Observe(elapsed)
		p.reqCnt.WithLabelValues(status, c.Request.Method+"_"+c.Request.URL.Path, c.HandlerName(), c.Request.Host, url).Inc()

	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
