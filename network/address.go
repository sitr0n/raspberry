package network

import ("io/ioutil"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

// Fix later
func create_port(index int) string {
	var port string = ":10000"
	s := strconv.Itoa(index)
	s = s_reverse(s)
	r_port := []rune(port)
	for i := len(port)-1; i >= 0; i-- {
		if(index < 10 && i < 5) {
			//fmt.Println(string(r_port))
			return string(r_port)
		} else if (index < 100 && i < 4) {
			//fmt.Println(string(r_port))
			return string(r_port)
		}
		number := []rune(s)[5-i]
		r_port[i] = number
	}
	return "0"
}

func s_reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}


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
	_REMOTES = len(addr_list)
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

func get_localip() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	check(err)
	defer conn.Close()

	ip_with_port := conn.LocalAddr().String()

	var ip string = ""
	for _, char := range ip_with_port {
		if (char == ':') {
			break
		}
		ip += string(char)
	}
	return ip
}
