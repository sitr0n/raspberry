package network

import ("io/ioutil"
	"encoding/json"
	"fmt"
)

func load_address() []remote_info {
	encoded, err := ioutil.ReadFile(ADDR_LIST_PATH)
	if (err != nil) {
		var addr []remote_info = []remote_info{}
		encoded, err := json.Marshal(addr)
		check(err)
		
		err = ioutil.WriteFile(ADDR_LIST_PATH, encoded, 0644)
		check(err)
		encoded, err = ioutil.ReadFile(ADDR_LIST_PATH)
	}
	
	var addr_list []remote_info
	err = json.Unmarshal(encoded, &addr_list)
	check(err)
	
	fmt.Println("Loaded address.")
	return addr_list
}

func save_address(addr_list []remote_info) {
	fmt.Println("Saving address list:", addr_list)
	encoded, err := json.Marshal(addr_list)
	check(err)

	err = ioutil.WriteFile(ADDR_LIST_PATH, encoded, 0644)
	check(err)
}

func remove_address(remove remote_info) []remote_info {
	list := load_address()
	n := len(list)
	temp_list := make([]remote_info, n)
	if (n > 0) {
		found := 0
		for i := 0; i < n; i++ {
			if (list[i].Name != remove.Name && list[i].Name != "") {
				temp_list[i - found] = list[i]
			} else {
				found ++
			}
		}
		if (found > 0) {
			list = make([]remote_info, n - found)
			copy(list, temp_list)
			fmt.Println("Removed:", remove.Name)
			save_address(list)
		} else {
			fmt.Println("Couldn't find referenced address.")
		}
	} else {
		fmt.Println("No addresses to remove.")
	}
	
	return list
}

func add_address(new_addr remote_info) []remote_info {
	list := load_address()
	n := len(list)
	name_check := true
	//fmt.Println("Perceived length:", n)
	for i := 0; i < n; i++ {
		if (new_addr.Name == list[i].Name) {
			name_check = false
		}
	}
	if (name_check == false) {
		fmt.Println("The remote name is not unique. Did not add the address.")
		return list
	}
	
	new_list := make([]remote_info, n+1)
	copy(new_list, list)
	list = new_list

	list = list[0 : n+1]
	list[n] = new_addr
	
	save_address(list)
	
	return list
}

func test_load() remote_info {
	var state = remote_info{}
 	var jsonBlob = []byte(`{"Name":"1","IP":"7"}`)
 	
 	err := json.Unmarshal(jsonBlob, &state)
 	check(err)
	
	return state
 }

