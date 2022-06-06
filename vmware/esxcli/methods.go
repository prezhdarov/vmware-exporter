package esxcli

import (
	"context"

	"github.com/vmware/govmomi/vim25/soap"
)

func (b *RetrieveManagedMethodExecuterBody) Fault() *soap.Fault { return b.Fault_ }

func (b *RetrieveDynamicTypeManagerBody) Fault() *soap.Fault { return b.Fault_ }

func (b *ExecuteSoapBody) Fault() *soap.Fault { return b.Fault_ }

func RetrieveDynamicTypeManager(ctx context.Context, r soap.RoundTripper, req *RetrieveDynamicTypeManagerRequest) (*RetrieveDynamicTypeManagerResponse, error) {
	var reqBody, resBody RetrieveDynamicTypeManagerBody

	reqBody.Req = req

	if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
		return nil, err
	}

	return resBody.Res, nil
}

func RetrieveManagedMethodExecuter(ctx context.Context, r soap.RoundTripper, req *RetrieveManagedMethodExecuterRequest) (*RetrieveManagedMethodExecuterResponse, error) {
	var reqBody, resBody RetrieveManagedMethodExecuterBody

	reqBody.Req = req

	if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
		return nil, err
	}

	return resBody.Res, nil
}

func ExecuteSoap(ctx context.Context, r soap.RoundTripper, req *ExecuteSoapRequest) (*ExecuteSoapResponse, error) {
	var reqBody, resBody ExecuteSoapBody

	reqBody.Req = req

	if err := r.RoundTrip(ctx, &reqBody, &resBody); err != nil {
		return nil, err
	}

	return resBody.Res, nil
}
