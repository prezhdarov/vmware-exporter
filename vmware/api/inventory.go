package vmware

/*
func (vm *VMware) inventory(loginData map[string]interface{}) error {

	// This is phase one - clusters and datastores. First are needed to grab the hosts with cluster <=> host relationship, the latter are just needed
	loginData["clusters"] = make(map[string]ClusterInventory)
	//	loginData["clusterList"] = []string{}
	loginData["datastores"] = make(map[string]DatastoreInventory)
	loginData["hosts"] = make(map[string]HostInventory)
	loginData["vms"] = make(map[string]VMInventory)

	errchan := make(chan error, 2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	for i := 0; i <= 1; i++ {

		go func(i int) {
			if i == 0 {
				vm.listClusters(loginData, errchan)
			} else {
				vm.listDatastores(loginData, errchan)
			}
			wg.Done()
		}(i)

	}

	for i := 0; i < 2; i++ {
		if err := <-errchan; err != nil {
			return err
		}
	}

	close(errchan)

	wg.Wait()

	err := vm.listHosts(loginData)
	if err != nil {
		return err
	}

	err = vm.listVMs(loginData)
	if err != nil {
		return err
	}

	return nil
}
*/
