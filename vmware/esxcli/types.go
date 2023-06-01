package esxcli

import (
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// The soap types......
type ReflectManagedMethodExecuterSoapArgument struct {
	types.DynamicData

	Name string `xml:"name"`
	Val  string `xml:"val"`
}

type ReflectManagedMethodExecuterSoapFault struct {
	types.DynamicData

	FaultMsg    string `xml:"faultMsg"`
	FaultDetail string `xml:"faultDetail,omitempty"`
}

type ReflectManagedMethodExecuterSoapResult struct {
	types.DynamicData

	Response string                                 `xml:"response,omitempty"`
	Fault    *ReflectManagedMethodExecuterSoapFault `xml:"fault,omitempty"`
}

type ExecuteSoapRequest struct {
	This     types.ManagedObjectReference               `xml:"_this"`
	Moid     string                                     `xml:"moid"`
	Version  string                                     `xml:"version"`
	Method   string                                     `xml:"method"`
	Argument []ReflectManagedMethodExecuterSoapArgument `xml:"argument,omitempty"`
}

type ExecuteSoapResponse struct {
	Returnval *ReflectManagedMethodExecuterSoapResult `xml:"urn:vim25 returnval"`
}

type ExecuteSoapBody struct {
	Req    *ExecuteSoapRequest  `xml:"urn:vim25 ExecuteSoap"`
	Res    *ExecuteSoapResponse `xml:"urn:vim25 ExecuteSoapResponse"`
	Fault_ *soap.Fault
}

// The DTM thingies
type InternalDynamicTypeManager struct {
	types.ManagedObjectReference
}

type RetrieveDynamicTypeManagerRequest struct {
	This types.ManagedObjectReference `xml:"_this"`
}

type RetrieveDynamicTypeManagerResponse struct {
	Returnval *InternalDynamicTypeManager `xml:"urn:vim25 returnval"`
}

type RetrieveDynamicTypeManagerBody struct {
	Req    *RetrieveDynamicTypeManagerRequest  `xml:"urn:vim25 RetrieveDynamicTypeManager"`
	Res    *RetrieveDynamicTypeManagerResponse `xml:"urn:vim25 RetrieveDynamicTypeManagerResponse"`
	Fault_ *soap.Fault
}

// The MME thingies
type ReflectManagedMethodExecuter struct {
	types.ManagedObjectReference
}

type RetrieveManagedMethodExecuterRequest struct {
	This types.ManagedObjectReference `xml:"_this"`
}

type RetrieveManagedMethodExecuterResponse struct {
	Returnval *ReflectManagedMethodExecuter `xml:"urn:vim25 returnval"`
}
type RetrieveManagedMethodExecuterBody struct {
	Req    *RetrieveManagedMethodExecuterRequest  `xml:"urn:vim25 RetrieveManagedMethodExecuter"`
	Res    *RetrieveManagedMethodExecuterResponse `xml:"urn:vim25 RetrieveManagedMethodExecuterResponse"`
	Fault_ *soap.Fault
}
