package pulsetictypes

//go:generate go tool enumer -type RequestType -trimprefix RequestType -json -text

//+kubebuilder:validation:Type:=string
//+kubebuilder:validation:Enum:=HTTP;TCP;ICMP

type RequestType uint8

const (
	RequestTypeHTTP RequestType = iota + 1
	RequestTypeTCP
	RequestTypeICMP
)
