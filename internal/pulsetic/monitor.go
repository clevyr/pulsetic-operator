package pulsetic

import (
	"github.com/clevyr/pulsetic-operator/internal/pulsetic/pulsetictypes"
)

type Monitor struct {
	ID                       int64                       `json:"id"`
	Name                     string                      `json:"name"`
	URL                      string                      `json:"url"`
	SSLCheck                 IntBool                     `json:"ssl_check"`
	IsRunning                bool                        `json:"is_running"`
	UptimeCheckFrequency     int                         `json:"uptime_check_frequency"`
	OfflineNotificationDelay int                         `json:"offline_notification_delay"`
	RequestType              pulsetictypes.RequestType   `json:"request_type"`
	TCPPorts                 string                      `json:"tcp_ports"`
	RequestMethod            pulsetictypes.RequestMethod `json:"request_method"`
	RequestBodyType          string                      `json:"request_body_type"`
	RequestBodyRaw           string                      `json:"request_body_raw"`
	RequestBodyJSON          string                      `json:"request_body_json"`
	RequestBodyFormParams    []FormParam                 `json:"request_body_form_params"`
	RequestHeaders           []Header                    `json:"request_headers"`
	RequestTimeout           float64                     `json:"request_timeout,omitzero"`
	ResponseBody             string                      `json:"response_body"`
	ResponseCode             string                      `json:"response_code"`
	ResponseHeaders          []Header                    `json:"response_headers"`
}

type FormParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
