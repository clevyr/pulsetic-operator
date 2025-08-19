package pulsetic

import (
	"github.com/clevyr/pulsetic-operator/internal/pulsetic/pulsetictypes"
)

type Monitor struct {
	ID                        int64                       `json:"id"`
	LatestCheckID             int64                       `json:"latest_check_id"`
	Name                      string                      `json:"name"`
	URL                       string                      `json:"url"`
	Status                    string                      `json:"status"`
	SSLCertificateState       string                      `json:"ssl_certificate_state"`
	SSLCheck                  IntBool                     `json:"ssl_check"`
	IP                        string                      `json:"ip"`
	Latitude                  float64                     `json:"latitude"`
	Longitude                 float64                     `json:"longitude"`
	IsRunning                 bool                        `json:"is_running"`
	IsNegative                int                         `json:"is_negative"`
	UptimeCheckFrequency      int                         `json:"uptime_check_frequency"`
	OfflineNotificationDelay  int                         `json:"offline_notification_delay"`
	RequestType               pulsetictypes.RequestType   `json:"request_type"`
	TCPPorts                  string                      `json:"tcp_ports"`
	RequestMethod             pulsetictypes.RequestMethod `json:"request_method"`
	RequestBodyType           string                      `json:"request_body_type"`
	RequestBodyRaw            string                      `json:"request_body_raw"`
	RequestBodyJSON           string                      `json:"request_body_json"`
	RequestBodyFormParams     []FormParam                 `json:"request_body_form_params"`
	RequestHeaders            []Header                    `json:"request_headers"`
	RequestTimeout            float64                     `json:"request_timeout"`
	ResponseBody              string                      `json:"response_body"`
	ResponseCode              string                      `json:"response_code"`
	ResponseHeaders           []Header                    `json:"response_headers"`
	ResponseExpectedCode      string                      `json:"response_expected_code"`
	SnapshotsAverageUptime    string                      `json:"snapshots_average_uptime"`
	Uptime                    float64                     `json:"uptime"`
	ResponseTime              float64                     `json:"response_time"`
	LighthouseAuditEnabled    bool                        `json:"lighthouse_audit_enabled"`
	MobilePerformanceScore    string                      `json:"mobile_performance_score"`
	MobileSEOScore            string                      `json:"mobile_seo_score"`
	MobilePWAScore            string                      `json:"mobile_pwa_score"`
	MobileAccessibilityScore  string                      `json:"mobile_accessibility_score"`
	MobileBestPracticesScore  string                      `json:"mobile_best_practices_score"`
	DesktopPerformanceScore   string                      `json:"desktop_performance_score"`
	DesktopSEOScore           string                      `json:"desktop_seo_score"`
	DesktopPWAScore           string                      `json:"desktop_pwa_score"`
	DesktopAccessibilityScore string                      `json:"desktop_accessibility_score"`
	DesktopBestPracticesScore string                      `json:"desktop_best_practices_score"`
	TotalChecks               int                         `json:"total_checks"`
	UnpaidChecks              int                         `json:"unpaid_checks"`
	BilledChecks              int                         `json:"billed_checks"`
	TelegramHash              string                      `json:"telegram_hash"`
	CheckedAt                 UnixOrTime                  `json:"checked_at"`
	Position                  int                         `json:"position"`
	Disabled                  int                         `json:"disabled"`
	StatusChangedAt           string                      `json:"status_changed_at"`
	FastChecked               int                         `json:"fast_checked"`
	CreatedAt                 string                      `json:"created_at"`
	UpdatedAt                 string                      `json:"updated_at"`
	DeletedAt                 string                      `json:"deleted_at"`
	SSLCertificate            SSLCertificate              `json:"ssl_certificate"`
	Nodes                     []Node                      `json:"nodes"`
}

type FormParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SSLCertificate struct {
	ID                 int64      `json:"id"`
	MonitorID          int64      `json:"monitor_id"`
	Domain             string     `json:"domain"`
	IssuedBy           string     `json:"issued_by"`
	SignatureAlgorithm string     `json:"signature_algorithm"`
	IsValid            IntBool    `json:"is_valid"`
	ExpiresAt          UnixOrTime `json:"expires_at"`
	CreatedAt          string     `json:"created_at"`
	UpdatedAt          string     `json:"updated_at"`
}

type Node struct {
	ID                  int64   `json:"id"`
	IP                  string  `json:"ip"`
	Location            string  `json:"location"`
	Title               string  `json:"title"`
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
	Available           int     `json:"available"`
	Status              string  `json:"status"`
	Uptime              int     `json:"uptime"`
	AverageResponseTime int     `json:"average_response_time"`
	LatestCheck         string  `json:"latest_check"`
	Active              bool    `json:"active"`
}
