package vmware

/*
type ClusterInventory struct {
	Name    string `json:"name"`
	Cluster string `json:"cluster"`
	HA      bool   `json:"ha_enabled"`
	DRS     bool   `json:"drs_enabled"`
}

func (vm *VMware) listClusters(loginData map[string]interface{}, er chan<- error) {

	var data []ClusterInventory

	extraConfig := make(map[string]interface{}, 0)

	extraConfig["api"] = "/api/vcenter/cluster"
	body, err := vm.Get(loginData, extraConfig)
	if err != nil {
		er <- err
		return
	}

	err = json.Unmarshal(*body.(*[]byte), &data)
	if err != nil {
		er <- err
		return
	}

	for _, cluster := range data {

		(loginData["clusters"].(map[string]ClusterInventory))[cluster.Cluster] = cluster

		//	loginData["clusterList"] = append(loginData["clusterList"].([]string), cluster.Cluster)
	}

	er <- nil
}
*/
