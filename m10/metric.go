package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var httpCost = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "httpserver",
		Name:      "http_cost",
		Help:      "http cost",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
	}, []string{"step"},
)

func initMetric() {
	prometheus.MustRegister(httpCost)
}

func setCost(beg time.Time) {
	httpCost.WithLabelValues("total").Observe(time.Now().Sub(beg).Seconds())
}
