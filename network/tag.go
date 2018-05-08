package network

import ("fmt"
	"math/rand"
)

func (r *Device) create_tag() tag {
	r.tag_req <- true
	granted := <- r.tag_grant
	return granted
}

func (r *Device) tag_handler() {
	
	var id_list []tag = []tag{}
	for { 
		select {
		case <- r.tag_req:
			new_tag := make_tag(&id_list)
			r.tag_grant <- new_tag
			
		case remove := <- r.tag_rm:
			id_list = remove_tag(id_list, remove)
			r.rm_list = remove_tag(r.rm_list, remove)
			
		case new_ack := <- r.ackchan:
			r.rm_list = add_tag(r.rm_list, new_ack)
			
		}	
	}
}

func make_tag(list *[]tag) tag {
	length := len(*list)
	list_copy := *list
	var id_unique bool
	var counter int
	var new_id tag
	for {
		counter ++
		new_id = random_tag()
		id_unique = true
		for i := 0; i < length; i++ {
			var check tag = list_copy[i]
			if (new_id == check) {
				id_unique = false
			}
		}
		if (id_unique == true) {
			break
		}
		if (counter > 10000) {
			fmt.Println("func make_tag is hanging. Tag list size:", length)
			counter = 0
		}
	}
	*list = add_tag(*list, new_id)
	
	return new_id
}

func random_tag() tag {
	random := rand.Intn(10000)
	var r tag = tag(random)
	return r
}

func add_tag(list []tag, new tag) []tag {
	n := len(list)
	if (n == cap(list)) {
		new_list := make([]tag, len(list), 2*len(list)+1)
		copy(new_list, list)
		list = new_list
	}
	list = list[0 : n+1]
	list[n] = new
	
	return list
}

func remove_tag(original []tag, remove tag) []tag {
	length := len(original)
	temp_list := make([]tag, length)
	if (length > 0) {
		found 	:= 0
		for i := 0; i < length; i++ {
			if (original[i] != remove ) {
				temp_list[i - found] = original[i]
			} else {
				found ++
			}
		}
		if (found > 0) {
			original = make([]tag, length - found)
			copy(original, temp_list)
		} else {
			//fmt.Println("Couldn't find referenced tag.", original, remove)
		}
	} else {
		//fmt.Println("No tags to remove.")
	}
	
	return original
}
