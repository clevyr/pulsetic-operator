package pulsetic

import (
	"strings"
)

type MonitorEditParams struct {
	URL                      string   `json:"url,omitzero"`
	Name                     string   `json:"name,omitzero"`
	UptimeCheckFrequency     int      `json:"uptime_check_frequency,string,omitzero"`
	OfflineNotificationDelay int      `json:"offline_notification_delay,string,omitzero"`
	SSLCheck                 IntBool  `json:"ssl_check,omitzero"`
	Request                  Request  `json:"request,omitzero"`
	Response                 Response `json:"response,omitzero"`
}

type Request struct {
	Type           string      `json:"type,omitzero"`
	BodyType       string      `json:"body_type,omitzero"`
	BodyRaw        string      `json:"body_raw,omitzero"`
	BodyJSON       string      `json:"body_json,omitzero"`
	BodyFormParams []FormParam `json:"body_form_params,omitzero"`
	Method         string      `json:"method,omitzero"`
	Headers        []Header    `json:"headers,omitzero"`
	Timeout        float64     `json:"timeout,omitzero"`
}

type Response struct {
	Body    string   `json:"body,omitzero"`
	Headers []Header `json:"headers,omitzero"`
}

func (m Monitor) EditParams() MonitorEditParams {
	return MonitorEditParams{
		URL:                      m.URL,
		Name:                     m.Name,
		UptimeCheckFrequency:     m.UptimeCheckFrequency,
		OfflineNotificationDelay: m.OfflineNotificationDelay,
		SSLCheck:                 m.SSLCheck,
		Request: Request{
			Type:           strings.ToLower(m.RequestType.String()),
			BodyType:       m.RequestBodyType,
			BodyRaw:        m.RequestBodyRaw,
			BodyJSON:       m.RequestBodyJSON,
			BodyFormParams: m.RequestBodyFormParams,
			Method:         strings.ToLower(m.RequestMethod.String()),
			Headers:        m.RequestHeaders,
			Timeout:        m.RequestTimeout,
		},
		Response: Response{
			Body:    m.ResponseBody,
			Headers: m.ResponseHeaders,
		},
	}
}
