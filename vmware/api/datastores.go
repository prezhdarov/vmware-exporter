package vmware

/*
type DatastoreInventory struct {
	Datastore string  `json:"datastore"`
	Name      string  `json:"name"`
	DSType    string  `json:"type"`
	Capacity  float64 `json:"capacity"`
	Free      float64 `json:"free_space"`
}

func (vm *VMware) listDatastores(loginData map[string]interface{}, er chan<- error) {

	var data []DatastoreInventory

	extraConfig := make(map[string]interface{}, 0)

	extraConfig["api"] = "/api/vcenter/datastore"
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

	for _, datastore := range data {

		(loginData["datastores"].(map[string]DatastoreInventory))[datastore.Datastore] = datastore

	}

	er <- nil

}
*/
