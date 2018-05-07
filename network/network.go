package network
import (
	_ "../config"
	addr "./addresses"
	. "../utility"
	"fmt"
	"net"
	"time"
	"encoding/json"
	//"os"
	"strconv"
	//"io/ioutil"
)

type ping struct {}

const(
	NO_GRANT	= -1

	_PING_PERIOD 	= 1000
	ADDR_LIST_PATH	= "../addresslist.json"
)

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
// Rename Remote
type RemoteList struct {
	remote		[]Remote
	store		addr.AddrList
	localip		string
}
// Rename Device
type Remote struct {
	Receive		chan interface{}
	Connected	chan bool
	id		int
	profile	  	addr.AddrData
	alive 		bool
	send		chan capsule
	tag_req		chan bool
	tag_rm		chan tag
	tag_grant	chan tag
	ackchan		chan tag
	rm_list		[]tag
}

var _localip string
func (r *RemoteList) Init() {

	var addrList addr.AddrList
	r.store = addrList
	r.store.Init("../data.json")
	r.localip = getLocalip()
	
	n := r.store.GetSize()
	r.remote = make([]Remote, n)
	for i := 0; i < n; i++ {
		r.remote[i].Receive		= make(chan interface{}, 100)
		r.remote[i].Connected 		= make(chan bool, 100)
		r.remote[i].id 			= i
		r.remote[i].profile		= r.store.GetData(i)	// init element for element?
		r.remote[i].alive 		= false
		r.remote[i].send 		= make(chan capsule, 100)
		r.remote[i].tag_req 		= make(chan bool, 100)
		r.remote[i].tag_rm 		= make(chan tag, 100)
		r.remote[i].tag_grant		= make(chan tag, 100)
		r.remote[i].ackchan 		= make(chan tag, 100)
		r.remote[i].rm_list		= []tag{}
		
		//go r.remote[i].tag_handler()
		//go r.remote[i].remote_listener(r.localip)
		//go r.remote[i].remote_broadcaster()
		//go r.remote[i].ping_remote()
	}
	
	go r.pairingListener()
}

func assertStr(d interface{}) string {
	if a_str, ok := d.(string); ok {
		return string(a_str)
	} else {
		fmt.Println("Something went wrong when sending ack.")
		return "0"
	}
}
/*
func (r *RemoteList) Add_remote(addr string) {
	// Replace load address, and expose the list for this func
	
	r.Receive		= make(chan interface{}, 100)
	r.Connected 		= make(chan bool, 100)
	r.id 			= r.store.GetSize()
	r.profile.IP		= addr
	//need to expose the damn list
	r.alive 		= false
	r.send 			= make(chan capsule, 100)
	r.tag_req 		= make(chan bool, 100)
	r.tag_rm 		= make(chan tag, 100)
	r.tag_grant		= make(chan tag, 100)
	r.ackchan 		= make(chan tag, 100)
	r.rm_list		= []tag{}
	
	//go r.tag_handler()
	//go r.remote_listener()
	//go r.remote_broadcaster()
	//go r.ping_remote()
	
	//add_address(new_addr remote_info) []remote_info
	//check_if_alive
	
	//initialize
	//share perceived port
	
}
*/
func (r *RemoteList) setupDevice(index int) Remote {
	var out Remote
	out.Receive		= make(chan interface{}, 100)
	out.Connected 		= make(chan bool, 100)
	out.id 			= index
	out.profile		= r.store.GetData(index)
	out.alive 		= false
	out.send 		= make(chan capsule, 100)
	out.tag_req 		= make(chan bool, 100)
	out.tag_rm 		= make(chan tag, 100)
	out.tag_grant		= make(chan tag, 100)
	out.ackchan 		= make(chan tag, 100)
	out.rm_list		= []tag{}
	return out
}

func (r *RemoteList) AddDevice(ip string) {
	n := len(r.remote)
	m := r.store.GetSize()
	Assert(n == m, "Desync between active- and saved remotes.")
	
	var newAddr addr.AddrData
	newAddr.IP = ip
	r.store.Add(newAddr)
	if (n < 1) {
		r.remote = make([]Remote, 1)
		r.remote[0] = r.setupDevice(0)
	} else {
		var add Remote
		add = r.setupDevice(n)
		tmpList := make([]Remote, n+1)
		copy(tmpList, r.remote)
		r.remote = tmpList
		r.remote = r.remote[0 : n+1]
		r.remote[n] = add
	}
	
	go r.remote[n].remote_listener(r.localip)
	
	r.requestPairing(r.remote[n].profile.IP, r.remote[n].profile.RPort)
	// start timer to remove store and remote if no response
	TPort := AssertInt(<- r.remote[n].Receive)
	
	r.store.SetTPort(n, TPort)
	r.remote[n].profile = r.store.GetData(n)
	
	go r.remote[n].remote_broadcaster()
	go r.remote[n].tag_handler()
	go r.remote[n].ping_remote()
}

func (r *RemoteList) GetSize() int {
	n := r.store.GetSize()
	return n
}

func (r *RemoteList) pairingListener() {
	listen_addr, err := net.ResolveUDPAddr("udp", r.localip + r.store.GetPairPort())
	check(err)
	in_connection, err := net.ListenUDP("udp", listen_addr)
	check(err)
	defer in_connection.Close()
	
	var message capsule
	
	for {
		buffer := make([]byte, 1024)
		length, _, _ := in_connection.ReadFromUDP(buffer)
		
		err := json.Unmarshal(buffer[:length], &message)
		check(err)
		
		fmt.Println("Received pairing request from:", in_connection.RemoteAddr)
	}
	
}

func (r *RemoteList) requestPairing(ip string, msg string) {
	target_addr,err := net.ResolveUDPAddr("udp", ip + r.store.GetPairPort())
	check(err)
	out_connection, err := net.DialUDP("udp", nil, target_addr)
	check(err)
	defer out_connection.Close()
	
	rPort, err := strconv.Atoi(msg)
	Assert(err == nil, "Pairing request failed to send")
	
	encoded, err := json.Marshal(rPort)
	check(err)
	out_connection.Write(encoded)
	//fmt.Println("Sent:", msg)
}

func (r *Remote) Send(idata interface{}) {
	var packet capsule = capsule{}
	
	switch DataType := idata.(type) {
	case ping:
		//fmt.Println("Sending ping!")
		packet.DataType = PING
		packet.ItemData= 0
		packet.ItemTag = 0
		r.send <- packet
	case ack:
		//fmt.Println("Sending ack!")
		packet.DataType = ACK
		packet.ItemData= assert_ack(idata)
		packet.ItemTag = r.create_tag()
		r.send <- packet	
	case int:
		//fmt.Println("Sending int!")
		packet.DataType = INT
		packet.ItemData= assert_int(idata)
		packet.ItemTag = r.create_tag()
		go r.sender(packet)
	case string:
		//fmt.Println("Sending string!")
		packet.DataType = STRING
		packet.ItemData= 0
		packet.ItemTag = r.create_tag()
		go r.sender(packet)
	default:
		fmt.Println("Unknown datatype!", DataType)
	}
}

func (r *Remote) sender(packet capsule) {
	
	miss := 0
	for {
		r.send <- packet
		time.Sleep(2*time.Second)
		r.tag_rm <- packet.ItemTag
		ok := r.checkForAck(packet.ItemTag)
		miss++
		if (ok == true || (miss < 3) || !r.alive) {
			//fmt.Println("DataAck chain complete!")
			break
		}
		packet.ItemTag = r.create_tag()
	}
}

func (r *Remote) checkForAck(t tag) bool {
	for _, e := range r.rm_list {
		if (t == e) {
			return true
		}
	}
	return false
}

func (r *Remote) remote_broadcaster() {
	target_addr,err := net.ResolveUDPAddr("udp", r.profile.IP + r.profile.TPort)
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

func (r *Remote) remote_listener(localip string) {
	listen_addr, err := net.ResolveUDPAddr("udp", localip + r.profile.RPort)
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
			r.Connected <- true
		}
		wd_kick <- true
		
		err := json.Unmarshal(buffer[:length], &message)
		check(err)
		//fmt.Println(length, "-----------",message)
		
		switch message.DataType {
		case PING:
			//fmt.Println("Received ping:", message)
		case ACK:
			//fmt.Println("Received ack:", int(message.ItemData))
			r.handle_ack(message.ItemData)
		case INT:
			//fmt.Println("Received int:", int(message.ItemData))
			r.send_ack(message)
			r.Receive <- message.ItemData
		case STRING:
			// ------------------------------------------------------------------ | TODO: Handle received strings
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
		r.Send(ping{})
	}
}

func (r *Remote) send_ack(reference capsule) {
	var response ack = ack{}
	response.ItemTag = item_tag2data(reference)
	r.Send(response)
}

// REDUNDANT - send ack capsule with datatag in itemtag
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
	r.Connected <- false
	fmt.Println("Connection with remote", r.id, "lost.")
}

func getLocalip() string {
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