package network
import cf "../config"
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

type ping struct {}

const(
	//PING 	ping	= false
	NO_GRANT	= -1

	_PING_PERIOD 	= 1000
	ADDR_LIST_PATH	= "../addresslist.json"
)

var PORT []string = []string {":10001",
				":10002"}
var IP []string = []string {"8.8.8.8",
				"9.9.9.9"}

type dtype int
const(
	PING	dtype	= 0
	ACK	dtype	= 1
	INT	dtype	= 2
	STRING	dtype	= 3
)
type tag int
type data int
type capsule struct {
	DataType	dtype
	ItemTag		tag
	ItemData	data
}

type ack struct {
	ItemTag		data
}

type Remote struct {
	id		int
	info	  	remote_info
	alive 		bool
	send		chan capsule
	received	chan interface{}
	ackchan		chan tag
	Reconnected	chan bool
	tag_req		chan bool
	tag_rm		chan tag
	tag_grant	chan tag	
}

type remote_info struct {
	Name		string
	IP		string
}
var _rm_list []tag // GÅR IKKE MED MER ENN TO REMOTES
var _localip string
var _REMOTES int
func Init(r *[cf.MAX_REMOTES]Remote) {
	_localip = get_localip()
	
	address_list := load_address()
	_REMOTES = len(address_list)
	
	for i := 0; i < _REMOTES; i++ {
		r[i].id 		= i
		r[i].info 		= address_list[i]
		r[i].alive 		= false
		r[i].send 		= make(chan capsule, 100)
		r[i].received 		= make(chan interface{}, 100)
		r[i].ackchan 		= make(chan tag, 100)
		r[i].Reconnected 	= make(chan bool, 100)
		r[i].tag_req 		= make(chan bool, 100)
		r[i].tag_rm 		= make(chan tag, 100)
		r[i].tag_grant		= make(chan tag, 100)
		
		go r[i].tag_handler()
		go r[i].remote_listener()
		go r[i].remote_broadcaster()
		go r[i].ping_remote()
	}
}

func (r *Remote) sender(packet capsule) {
	
	miss := 0
	for {
		r.send <- packet
		time.Sleep(2*time.Second)
		r.tag_rm <- packet.ItemTag
		ok := r.check_for_ack(packet.ItemTag)
		miss++
		if (ok == true || (miss < 3)) {
			fmt.Println("DataAck chain complete!")
			break
		}
		packet.ItemTag = r.create_tag()
	}
}

func (r *Remote) check_for_ack(t tag) bool {
	for _, e := range _rm_list {
		if (t == e) {
			return true
		}
	}
	return false
}


func (r *Remote) remote_broadcaster() {
	target_addr,err := net.ResolveUDPAddr("udp", r.info.IP + PORT[r.id])
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
			//fmt.Println("Sent:", msg)
		}
	}
}





func (r *Remote) remote_listener() {
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + PORT[r.id])
	check(err)
	in_connection, err := net.ListenUDP("udp", listen_addr)
	check(err)
	defer in_connection.Close()
	//var ack ack
	
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
		
		err := json.Unmarshal(buffer[:length], &message)
		check(err)
		//fmt.Println(length, "-----------",message)
		
		switch message.DataType {
		case PING:
			fmt.Println("Received ping:", message)
		case ACK:
			fmt.Println("Received ack:", int(message.ItemData))
			r.handle_ack(message.ItemData)
		case INT:
			fmt.Println("Received int:", int(message.ItemData))
			r.send_ack(message)
		case STRING:
			fmt.Println("Received string:", string(message.ItemData))
		default:
			fmt.Println("Received data:", message.ItemData)
		}
		
	}
}

func (r *Remote) handle_ack(d data) {
	//r.tag_rm <- tag(d)
	r.ackchan <- tag(d)
}
func (r *Remote) ping_remote() {
	const active	time.Duration = time.Duration(_PING_PERIOD)*time.Millisecond
	const idle 	time.Duration = 5*time.Second
	//var p = ping{}

	for {
		if (r.alive) {
			time.Sleep(active)
		} else {
			time.Sleep(idle)
		}
		r.Send(int(12))
		//r.Send(ping{})
	}
}

func (r *Remote) send_ack(reference capsule) {
	var response ack = ack{}
	response.ItemTag = item_tag2data(reference)
	r.Send(response)
}

func item_tag2data(it capsule) data {
	var re data
	re = assert_capsule(it)
	return re
}

func assert_capsule(d interface{}) data {
	
	if a_int, ok := d.(capsule); ok {
		var idata data
		idata = data(a_int.ItemTag)
		return idata
	} else {
		fmt.Println("Something went wrong asserting capsule.")
		return 0
	}
}


func (r *Remote) Send(idata interface{}) {
	var packet capsule = capsule{}
	
	switch DataType := idata.(type) {
	case ping:
		fmt.Println("Sending ping!")
		packet.DataType = PING
		packet.ItemData= 0
		packet.ItemTag = 0
		r.send <- packet
	case ack:
		fmt.Println("Sending ack!")
		packet.DataType = ACK
		packet.ItemData= assert_ack(idata)
		packet.ItemTag = r.create_tag()
		r.send <- packet	
	case int:
		fmt.Println("Sending int!")
		packet.DataType = INT
		packet.ItemData= assert_int(idata)
		packet.ItemTag = r.create_tag()
		go r.sender(packet)
	case string:
		fmt.Println("Sending string!")
		packet.DataType = STRING
		packet.ItemData= 0
		packet.ItemTag = r.create_tag()
		go r.sender(packet)
	default:
		fmt.Println("Unknown datatype!", DataType)
	}
}

func assert_ack(d interface{}) data {
	if a_int, ok := d.(ack); ok {
		return data(a_int.ItemTag)
	} else {
		fmt.Println("Something went wrong when sending ack.")
		return 0
	}
}

func assert_int(d interface{}) data {
	if a_int, ok := d.(int); ok {
		return data(a_int)
	} else {
		fmt.Println("Something went wrong when sending int.")
		return 0
	}
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



func (r *Remote) create_tag() tag {
	r.tag_req <- true
	granted := <- r.tag_grant
	return granted
}

func (r *Remote) tag_handler() {
	fmt.Println("R", r.id, "- tag handler started.")
	var id_list []tag = []tag{}
	 _rm_list = []tag{}
	for { 
		select {
		case <- r.tag_req:
			new_tag := make_tag(&id_list)
			r.tag_grant <- new_tag
			
		case remove := <- r.tag_rm:
			id_list = remove_tag(id_list, remove)
			_rm_list = remove_tag(_rm_list, remove)
			
		case new_ack := <- r.ackchan:
			_rm_list = add_tag(_rm_list, new_ack)
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

func Get_remotes() int {
	return _REMOTES
}
