package esxcli

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
)

func Run(ctx context.Context, client *vim25.Client, host types.ManagedObjectReference, command []string, arguments map[string]string) (interface{}, error) {

	//{
	res, err := GetHostMME(ctx, client, &host)
	if err != nil {

		return nil, err
	}

	request := ExecuteSoapRequest{
		This:     *res,
		Moid:     "ha-cli-handler-" + strings.Join(command[:len(command)-1], "-"),
		Method:   "vim.EsxCLI." + strings.Join(command, "."),
		Version:  "urn:vim25/5.0",
		Argument: ConfigArguments(arguments),
	}

	x, err := ExecuteSoap(ctx, client, &request)
	if err != nil {
		return nil, err
	}

	if x.Returnval != nil {
		if x.Returnval.Fault != nil {
			return nil, errors.New(x.Returnval.Fault.FaultMsg)
		}

	}

	//}

	//This remains here in case we need DTR :D
	/*	{
		req := RetrieveDynamicTypeManagerRequest{
			This: host,
		}

		res, err := RetrieveDynamicTypeManager(ctx, client, &req)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			//return nil, err
		}

		dtm := res.Returnval

		fmt.Printf("dtm: %s\n", dtm)
	} */

	return x.Returnval.Response, nil
}

func ConfigArguments(args map[string]string) []ReflectManagedMethodExecuterSoapArgument {

	var sargs []ReflectManagedMethodExecuterSoapArgument

	for argname, argvalue := range args {

		sargs = append(sargs, ReflectManagedMethodExecuterSoapArgument{Name: argname, Val: fmt.Sprintf("<%s>%s</%s>", argname, argvalue, argname)})
	}

	return sargs

}

func GetHostMME(ctx context.Context, client *vim25.Client, host *types.ManagedObjectReference) (*types.ManagedObjectReference, error) {

	req := RetrieveManagedMethodExecuterRequest{
		This: *host,
	}

	//{
	res, err := RetrieveManagedMethodExecuter(ctx, client, &req)
	if err != nil {

		return &types.ManagedObjectReference{}, err
	}

	return &res.Returnval.ManagedObjectReference, nil
}

func GetSOAP(ctx context.Context, client *vim25.Client, request *ExecuteSoapRequest, data interface{}) error {
	res, err := ExecuteSoap(ctx, client, request)
	if err != nil {
		return fmt.Errorf("error executing soap request: %s", err)
	}

	if res.Returnval != nil {
		if res.Returnval.Fault != nil {
			return fmt.Errorf("error at xml return value: %s", err)
		}

	}

	err = xml.Unmarshal([]byte(res.Returnval.Response), &data)
	if err != nil {
		return fmt.Errorf("error unmarshalling xml: %s", err)
	}

	return nil
}
