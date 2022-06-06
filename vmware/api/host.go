package vmware

/*
type HostInventory struct {
	Host    string `json:"host"`
	Name    string `json:"name"`
	Cluster string //No tag here, this is going to be used for cluster parent reference
}

func (vm *VMware) listHosts(loginData map[string]interface{}) error {

	clusterCount := len(loginData["clusters"].(map[string]ClusterInventory))

	errchan := make(chan error, clusterCount)

	wg := sync.WaitGroup{}
	wg.Add(clusterCount)

	for _, clusterData := range loginData["clusters"].(map[string]ClusterInventory) {

		go func(clusterData ClusterInventory) {
			vm.getClusterHosts(clusterData, loginData, errchan)
			wg.Done()
		}(clusterData)

	}

	for i := 0; i < clusterCount; i++ {
		if err := <-errchan; err != nil {
			return fmt.Errorf("error here %s", err)
		}
	}

	close(errchan)

	wg.Wait()

	return nil
}

func (vm *VMware) getClusterHosts(cluster ClusterInventory, loginData map[string]interface{}, ec chan<- error) {

	var data []HostInventory

	extraConfig := make(map[string]interface{}, 0)

	extraConfig["api"] = fmt.Sprintf("/api/vcenter/host?clusters=%s&connection_states=CONNECTED", cluster.Cluster)
	body, err := vm.Get(loginData, extraConfig)
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

	for _, host := range data {

		host.Cluster = cluster.Name

		(loginData["hosts"].(map[string]HostInventory))[host.Host] = host

	}

	ec <- nil
}
*/
