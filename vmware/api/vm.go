package vmware

/*
type VMInventory struct {
	VM     string  `json:"vm"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu_count"`
	Memory float64 `json:"memory_size_MiB"`
	Host   string  //No tag here, this is going to be used for cluster parent reference
}

func (vm *VMware) listVMs(loginData map[string]interface{}) error {

	hostCount := len(loginData["hosts"].(map[string]HostInventory))

	errchan := make(chan error, hostCount)

	wg := sync.WaitGroup{}
	wg.Add(hostCount)

	for _, hostData := range loginData["hosts"].(map[string]HostInventory) {

		go func(hostData HostInventory) {
			vm.getHostVMs(hostData, loginData, errchan)
			wg.Done()
		}(hostData)

	}

	for i := 0; i < hostCount; i++ {
		if err := <-errchan; err != nil {
			return fmt.Errorf("error here %s", err)
		}
	}

	close(errchan)

	wg.Wait()

	return nil
}

func (vmw *VMware) getHostVMs(host HostInventory, loginData map[string]interface{}, ec chan<- error) {

	var data []VMInventory

	extraConfig := make(map[string]interface{}, 0)

	extraConfig["api"] = fmt.Sprintf("/api/vcenter/vm?hosts=%s&power_states=POWERED_ON", host.Host)
	body, err := vmw.Get(loginData, extraConfig)
	if err != nil {
		ec <- err
		return
	}

	err = json.Unmarshal(*body.(*[]byte), &data)
	if err != nil {
		ec <- fmt.Errorf("error: %s", string(*body.(*[]byte)))
		return
	}

	ldWriteMtx.Lock()
	defer ldWriteMtx.Unlock()

	for _, vm := range data {

		vm.Host = host.Name

		(loginData["vms"].(map[string]VMInventory))[vm.VM] = vm

	}

	ec <- nil
}
*/
