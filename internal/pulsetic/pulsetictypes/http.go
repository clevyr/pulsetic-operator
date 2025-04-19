package pulsetictypes

//go:generate go run github.com/dmarkham/enumer -type RequestMethod -trimprefix Method -json -text

//+kubebuilder:validation:Type:=string
//+kubebuilder:validation:Enum:=GET;POST;PUT;PATCH;DELETE;HEAD;OPTIONS

type RequestMethod uint8

const (
	MethodGET RequestMethod = iota
	MethodPOST
	MethodPUT
	MethodPATCH
	MethodDELETE
	MethodHEAD
	MethodOPTIONS
)
