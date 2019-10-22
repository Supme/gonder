package models

import "github.com/prometheus/client_golang/prometheus"

var Prometheus struct {
	Api struct {
		AuthRequest prometheus.Counter
		UserRequest *prometheus.CounterVec
	}
	Campaign struct{
		Started     prometheus.Gauge
		SendResult *prometheus.CounterVec
	}
	UTM struct {
		Request *prometheus.CounterVec
	}
}

func InitPrometheus() {
	Prometheus.Api.AuthRequest = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gonder_api_auth_request",
		})
	Prometheus.Api.UserRequest = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gonder_api_user_request",
		},[]string{"ip"})
	Prometheus.Campaign.Started = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gonder_campaign_started",
		})
	Prometheus.Campaign.SendResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gonder_campaign_send_result",
		},[]string{"campaign", "domain", "status", "type"})

	Prometheus.UTM.Request = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gonder_utm_request",
		},[]string{"type"})

	prometheus.MustRegister(
		Prometheus.Api.AuthRequest,
		Prometheus.Api.UserRequest,
		Prometheus.Campaign.Started,
		Prometheus.Campaign.SendResult,
		Prometheus.UTM.Request,
		)
}