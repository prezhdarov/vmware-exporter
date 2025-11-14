package vmware

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/prezhdarov/prometheus-exporter/collector"

	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

var (
	vmwUser       = flag.String("vmware.username", "", "Username to login to vCenter server")
	vmwPasswd     = flag.String("vmware.password", "", "Password for the user above")
	vCenter       = flag.String("vmware.vcenter", "", "vCenter server address in host:port format. This is not the vCenter Management Console")
	vmwSchema     = flag.String("vmware.schema", "https", "Use HTTP or HTTPS")
	vmwTLS        = flag.Bool("vmware.insecureTLS", false, "Trust insecure vCenter TLS (true) or verify (default)")
	vmwInterval   = flag.Int("vmware.interval", 20, "How often data will be collected. Default is every 20s.")
	vmGranularity = flag.Int("vmware.granularity", 20, "The frequency of the sampled data. Default is 20s")

	//ldWriteMtx = sync.Mutex{}
)

type VMware struct {
	//logger log.Logger
}

func init() {

	collector.RegisterAPI(NewAPI())

}

func NewAPI() *VMware {

	return &VMware{}
}

func Load(logger *slog.Logger) {

	logger.Info("msg", "Loading VMware vSphere API", nil)

}

func (vm *VMware) Login(target string, logger *slog.Logger) (map[string]interface{}, error) {

	loginData := make(map[string]interface{}, 0)

	if target == "" {

		target = *vCenter

	}

	loginData["target"] = target

	//Login into REST API (get session key) - this is getting parked for now
	//if err := restLogin(loginData); err != nil {
	//	return nil, err
	//}

	//Login into REST API (get session key)
	if err := govmomiLogin(loginData); err != nil {
		return nil, err
	}

	//Fill in the REST inventory
	//if err := vm.inventory(loginData); err != nil {

	//	return nil, err

	//}

	//Here we login using govmomi (I guess)

	return loginData, nil
}

func (vm *VMware) Logout(loginData map[string]interface{}, logger *slog.Logger) error {

	/*
		url := fmt.Sprintf("%s://%s/api/session", *vmwSchema, loginData["target"].(string))

		statusCode, _, body, err := request("DELETE", url, loginData["headers"].(map[string]string), true)
		if err != nil {
			return err
		}

		if statusCode != 204 {
			return fmt.Errorf("Login failed with status code: %d, and body %s", statusCode, body)
		}
	*/

	return nil

}

/*func (vm *VMware) Get(loginData, extraConfig map[string]interface{}) (interface{}, error) {

	if extraConfig["type"].(string) == "manager" {

		return vm.restGet(loginData["target"].(string), loginData["session"].(string), loginData["headers"].(map[string]string))

	} else if extraConfig["type"].(string) == "gateway" {

		return f.GetGateway(loginData["target"].(string), loginData["session"].(string), extraConfig["api"].(string), extraConfig["gateways"].(string))
	}

	return nil, fmt.Errorf("wrong or undefined target type")
}*/

func (vm *VMware) Get(loginData, extraConfig map[string]interface{}, logger *slog.Logger) (interface{}, error) {

	url := fmt.Sprintf("%s://%s%s", *vmwSchema, loginData["target"], extraConfig["api"])

	_, _, body, err := request("GET", url, loginData["headers"].(map[string]string), false)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

// request is where the http magic happens
func request(method, url string, headers map[string]string, login bool) (int, string, []byte, error) {

	transport := http.DefaultTransport
	transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: *vmwTLS}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(*vmwInterval-2) * time.Second,
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, "", nil, err
	}

	if login {
		req.SetBasicAuth(*vmwUser, *vmwPasswd)
	}

	for header := range headers {
		req.Header.Add(header, headers[header])
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", nil, err
	}

	responseHeaders := resp.Header.Get("cookie")

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return 0, "", nil, err
	}

	return resp.StatusCode, responseHeaders, body, nil
}

/*
func restLogin(loginData map[string]interface{}) error {

	url := fmt.Sprintf("%s://%s/api/session", *vmwSchema, loginData["target"].(string))
	headers := map[string]string{"Content-type": "application/json"}

	statusCode, _, body, err := request("POST", url, headers, true)
	if err != nil {

		return err

	}

	if statusCode != 201 {

		return fmt.Errorf("Login failed with status code: %d, and body %s", statusCode, body)

	}

	headers["vmware-api-session-id"] = string(body)[1:(len(body) - 1)]
	loginData["headers"] = headers

	return nil
}
*/

func govmomiLogin(loginData map[string]interface{}) error {

	//Prep the url for SOAP login
	urlx, err := soap.ParseURL(fmt.Sprintf("%s://%s%s", *vmwSchema, loginData["target"].(string), vim25.Path))
	if err != nil {
		return fmt.Errorf("soap url err: %s", err)
	}

	urlx.User = url.UserPassword(*vmwUser, *vmwPasswd)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*vmwInterval-2)*time.Second)

	session := &cache.Session{URL: urlx, Insecure: *vmwTLS, Passthrough: true}

	client := new(vim25.Client)

	err = session.Login(ctx, client, nil)
	if err != nil {
		cancel()
		return fmt.Errorf("login err: %s", err)
	}

	//Property spec Manager
	loginData["view"] = view.NewManager(client)

	//Performance Manager and performance counters
	loginData["perf"] = performance.NewManager(client)
	loginData["counters"], err = loginData["perf"].(*performance.Manager).CounterInfoByName(ctx)
	if err != nil {
		cancel()
		return fmt.Errorf("perfman counters err: %s", err)
	}

	loginData["cancel"] = cancel

	//Finally add the context and the govmomi client itself
	loginData["ctx"] = ctx
	loginData["client"] = client
	loginData["interval"] = int32(*vmwInterval)
	loginData["samples"] = int32(*vmwInterval / *vmGranularity)

	return nil
}
