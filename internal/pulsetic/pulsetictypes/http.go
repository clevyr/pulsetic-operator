package pulsetictypes

//go:generate go tool enumer -type RequestMethod -trimprefix Method -json -text

//+kubebuilder:validation:Type:=string
//+kubebuilder:validation:Enum:=GET;POST;PUT;PATCH;DELETE;HEAD;OPTIONS

type RequestMethod uint8

const (
	MethodHEAD RequestMethod = iota
	MethodGET
	MethodPOST
	MethodPUT
	MethodPATCH
	MethodDELETE
	MethodOPTIONS
)
