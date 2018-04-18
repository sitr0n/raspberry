package network
import (
	"fmt"
	"net"
	"time"
	"encoding/json"
	//"os"
	//"strconv"
	"math/rand"
	//"io/ioutil"
)

const(
	PING 		= false
	Ack_order 	= 0
	Ack_state 	= 1
	Ack_order_accept= 2
	NO_GRANT	= -1

	_PING_PERIOD 	= 1000
	ADDR_LIST_PATH	= "../addresslist.json"
)

var PORT []string = []string {":10001",
				":10002"}
var IP []string = []string {"8.8.8.8",
				"9.9.9.9"}

type tag int
type capsule struct {
	item_tag	tag
	data		interface{}
}

type ack struct {
	item_tag	tag
}

type Remote struct {
	id		int
	address  	string
	alive 		bool
	send		chan interface{}
	received	chan interface{}
	ackchan		chan ack
	Reconnected	chan bool
	tag_req		chan bool
	tag_rm		chan tag
	tag_grant	chan tag	
}

type remote_info struct {
	Name		string
	IP		string
}

var _localip string

func Init(r *[2]Remote) {
	_localip = get_localip()
	
	address_list := load_address()
	NUM_REMOTES := len(address_list)
	
	for i := 0; i < NUM_REMOTES; i++ {
		r[i].id 		= i
		//r[i].address 		= address_list[i].IP
		r[i].alive 		= false
		r[i].send 		= make(chan interface{}, 100)
		r[i].received 		= make(chan interface{}, 100)
		r[i].ackchan 		= make(chan ack, 100)
		r[i].Reconnected 	= make (chan bool, 100)
		r[i].tag_req 		= make (chan bool, 100)
		r[i].tag_rm 		= make(chan tag, 100)
		r[i].tag_grant		= make(chan tag, 100)
		
		//go r[i].tag_handler()
		//go r[i].remote_listener()
		//go r[i].remote_broadcaster()
	}
	//go ping_remotes(r)
	
}

func (r *Remote) await_ack(expecting int) bool {
	timeout := make(chan bool)
	timer_cancel := make(chan bool)
	go timeout_timer(timer_cancel, timeout)
	
	select {
	case <- r.ackchan:
		timer_cancel <- true
		return true
	
	case <- timeout:
		return false
	}
}

func (r *Remote) send_ack(reference int) {
	var response ack = ack{}
	response.item_tag = tag(reference)
	r.send <- response
}

func (r *Remote) remote_listener() {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + PORT[r.id])
	check(err)
	in_connection, err := net.ListenUDP("udp", listen_addr)
	check(err)
	defer in_connection.Close()
	
	//var ack ack

	const ACK_SIZE = 2
	
	var message capsule
	
	wd_kick := make(chan bool, 100)
	for {
		buffer := make([]byte, 1024)
		length, _, _ := in_connection.ReadFromUDP(buffer)
		if (r.alive == false) {
			go r.watchdog(wd_kick)
			fmt.Println("Connection with remote", r.id, "established!")
			r.Reconnected <- true
		}
		wd_kick <- true
		
		switch length {
		case ACK_SIZE:
			fmt.Println("Received ack!")
		default:
			
			err := json.Unmarshal(buffer[:length], &message)
			check(err)
			fmt.Println("Received package!\nSize:", length)
			fmt.Println("Tag:", message.item_tag, "\nData:", message.data)
		}
	}
}

func (r *Remote) remote_broadcaster() {
	target_addr,err := net.ResolveUDPAddr("udp", r.address + PORT[r.id])
	check(err)
	out_connection, err := net.DialUDP("udp", nil, target_addr)
	check(err)
	defer out_connection.Close()

	for {
		select {
		case msg := <- r.send:
			encoded, err := json.Marshal(msg)
			check(err)
			out_connection.Write(encoded)
			fmt.Println("Sent:", msg)
		}
	}
}

func ping_remotes(remote *[1]Remote) { // 2 LITERALS INSTEAD OF CONST
	for {
		time.Sleep(time.Duration(_PING_PERIOD)*time.Millisecond)
		for i := 0; i < 1; i++ {
			remote[i].send <- PING
		}
	}
}

func timeout_timer(cancel <- chan bool, timeout chan <- bool) {
	for i := 0; i < 10; i++ {
		time.Sleep(500*time.Millisecond)
		select {
		case <- cancel:
			return

		default:
		}
	}
	timeout <- true
}

func flush_channel(c <- chan interface{}) {
	for i := 0; i < 100; i++ {
		select {
		case <- c:
		default:
		}
	}
}

func (r *Remote) watchdog(kick <- chan bool) {
	r.alive = true
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(_PING_PERIOD)*time.Millisecond)
		select {
		case <- kick:
			i = 0
		default:
		}
	}
	r.alive = false
	fmt.Println("Connection with remote", r.id, "lost.")
}

/*
func ip_address(adr string) string {
	var is_int = true
	for _, char := range adr {
		if (char == '.') {
			is_int = false
		}
	}
	if (is_int == true) {
		index, err := strconv.Atoi(adr)
		if (err != nil) {
			fmt.Println("Argument is invalid, try another.")
			os.Exit(2)
		}
		if (index > def.WORKSPACES || index < 1) {
			fmt.Println("An argument is out of bounds. Please try another number or target IP address.")
			os.Exit(2)
		}
		return def.WORKSPACE[index]
	} else {
		return adr
	}
}
*/

func (r *Remote) Send(data interface{}) {
	
	var packet capsule = capsule{}
	packet.data 	= data
	packet.item_tag = r.create_tag()
	
	r.send <- packet
}

func (r *Remote) create_tag() tag {
	r.tag_req <- true
	granted := <- r.tag_grant
	return granted
}

func (r *Remote) tag_handler() {
	fmt.Println("R", r.id, "- tag handler started.")
	var id_list []tag = []tag{}
	for{ 
		select {
		case <- r.tag_req:
			new_tag := make_tag(&id_list)
			r.tag_grant <- new_tag
			
		case remove := <- r.tag_rm:
			id_list = remove_tag(id_list, remove)
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
			//original = make([]tag, length - found)
			copy(original, temp_list)
		} else {
			fmt.Println("Couldn't find referenced tag.")
		}
	} else {
		fmt.Println("No tags to remove.")
	}
	
	return original
}

/*
func ip_address(adr interface{}) def.IP {
	switch a := adr.(type) {
	case string:
		return a
	case int:
		if (a > 23 || a < 0) {
			fmt.Println("Workspace index is out of bounds. Please abort process and try another argument!")
			for {
			}
		} else {
			return def.WORKSPACE[a]
		}
	default:
		fmt.Println("Wrong data type passed to network.Init. Try string or workspace number.")
		return "0"
	}
}
*/

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



func check(e error) {
	if (e != nil) {
		panic(e)
	}
}
