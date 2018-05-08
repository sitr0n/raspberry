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
	PAIRING dtype	= 4
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
	device		[]Device
	store		addr.AddrList
	localip		string
}

type Device struct {
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
func (r *Remote) Init() {

	var addrList addr.AddrList
	r.store = addrList
	r.store.Init("../data.json")
	r.localip = getLocalip()
	
	n := r.store.GetSize()
	r.device = make([]Device, n)
	for i := 0; i < n; i++ {
		r.device[i].Receive		= make(chan interface{}, 100)
		r.device[i].Connected 		= make(chan bool, 100)
		r.device[i].id 			= i
		r.device[i].profile		= r.store.GetData(i)	// init element for element?
		r.device[i].alive 		= false
		r.device[i].send 		= make(chan capsule, 100)
		r.device[i].tag_req 		= make(chan bool, 100)
		r.device[i].tag_rm 		= make(chan tag, 100)
		r.device[i].tag_grant		= make(chan tag, 100)
		r.device[i].ackchan 		= make(chan tag, 100)
		r.device[i].rm_list		= []tag{}
		
		//go r.device[i].tag_handler()
		//go r.device[i].remote_listener(r.localip)
		//go r.device[i].remote_broadcaster()
		//go r.device[i].ping_remote()
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

func (r *Remote) setupDevice(index int) Device {
	var out Device
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

func (r *Remote) Add(ip string) {
	n := len(r.device)
	r.addDevice(ip)
	fmt.Println("add stored")
	go r.device[n].remote_listener(r.localip)
	time.Sleep(time.Second)
	r.requestPairing(r.device[n].profile.IP, r.device[n].profile.RPort)
	fmt.Println("add requested")
	
	timeout := make(chan bool)
	cancel := make(chan bool)
	go Timer(1000, cancel, timeout)
	select {
	case <- timeout:
		fmt.Println("Pairing failed!")
		r.store.Remove(r.device[n].profile)	// ----------------------------------------- Replace this with local function which removes the device struct as well
	
	case response := <- r.device[n].Receive:
		fmt.Println("pairing data landed in Add func")
		cancel <- true
		
		TPort := AssertInt(response)
		fmt.Println("tport received")
		r.setTPort(n, TPort)
		go r.device[n].remote_broadcaster()

		var packet capsule = capsule{}
		packet.DataType = PAIRING
		packet.ItemData = data(420)
		r.device[n].send <- packet
		
		go r.device[n].tag_handler()
		go r.device[n].ping_remote()
		fmt.Println("Pairing complete!", response)
	}
	
	
	
}

func (r *Remote) addDevice(ip string) {
	n := len(r.device)
	m := r.store.GetSize()
	Assert(n == m, "Desync between active- and saved remotes.")
	
	r.store.Add(ip)
	if (n < 1) {
		r.device = make([]Device, 1)
		r.device[0] = r.setupDevice(0)
	} else {
		var add Device
		add = r.setupDevice(n)
		tmpList := make([]Device, n+1)
		copy(tmpList, r.device)
		r.device = tmpList
		r.device = r.device[0 : n+1]
		r.device[n] = add
	}
}

func (r *Remote) pairingListener() {
	listen_addr, err := net.ResolveUDPAddr("udp", r.localip + r.store.GetPairPort())
	Check(err)
	in_connection, err := net.ListenUDP("udp", listen_addr)
	Check(err)
	defer in_connection.Close()
	
	var message capsule
	
	for {
		buffer := make([]byte, 1024)
		length, adr, _ := in_connection.ReadFromUDP(buffer)
		
		err := json.Unmarshal(buffer[:length], &message)
		Check(err)
		IP := stripPort(adr.String())
		ok := r.checkIP(IP)
		if (ok) {
			fmt.Println("New IP found!")
			var msg interface{}
			msg = message.ItemData
			var port int = int(msg.(data))
			
			n := r.store.GetSize()
			r.addDevice(IP)
			
			r.setTPort(n, port)
			fmt.Println("New device stored!")
			
			go r.device[n].remote_broadcaster()

			rport, _ := strconv.Atoi(r.device[n].profile.RPort)
			
			var packet capsule = capsule{}
			packet.DataType = PAIRING
			packet.ItemData = data(rport)
			r.device[n].send <- packet
			
			timeout := make(chan bool)
			cancel := make(chan bool)
			go Timer(1000, cancel, timeout)
			
			select {
			case <- timeout:
				fmt.Println("Pairing failed!")
				r.store.Remove(r.device[n].profile)	// ----------------------------------------- Replace this with local function which removes the device struct as well
			
			case <- r.device[n].Receive:
				cancel <- true
				go r.device[n].tag_handler()
				go r.device[n].ping_remote()
				go r.device[n].remote_listener(r.localip)
			}
		}
	}
}



func (r *Remote) Reset() {
	fmt.Println("Resetting...")
	r.device = nil
	r.store.ClearAllData()
}

func (r *Remote) setTPort(index int, port int) {
	r.store.SetTPort(index, port)
	r.device[index].profile = r.store.GetData(index)
	fmt.Println("After setting tport:", r.device[index].profile.TPort)
}

func (r *Remote) checkIP(address string) bool {
	check := true
	for _, dev := range r.device {
		if (dev.profile.IP == address) {
			check = false
		}
	}
	return check
}

func (r *Remote) requestPairing(ip string, msg string) {
	fmt.Println("requestPairing:", ip + r.store.GetPairPort())
	target_addr,err := net.ResolveUDPAddr("udp", ip + r.store.GetPairPort())
	Check(err)
	out_connection, err := net.DialUDP("udp", nil, target_addr)
	Check(err)
	defer out_connection.Close()
	
	rPort, err := strconv.Atoi(msg)
	Assert(err == nil, "Pairing request failed to send")
	
	var packet capsule = capsule{}
	packet.DataType = PAIRING
	packet.ItemData = data(rPort)
	
	encoded, err := json.Marshal(packet)
	Check(err)
	out_connection.Write(encoded)
	//fmt.Println("Sent:", msg)
}

func (r *Device) Send(idata interface{}) {
	var packet capsule = capsule{}
	fmt.Println("Sending", idata, "from device id:", r.id)
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
		fmt.Println("Sending int!")
		packet.DataType = INT
		packet.ItemData= assert_int(idata)
		packet.ItemTag = r.create_tag()
		fmt.Println("sender starting")
		go r.sender(packet)
		fmt.Println("sender started")
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

func (r *Device) sender(packet capsule) {
	
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

func (r *Device) checkForAck(t tag) bool {
	for _, e := range r.rm_list {
		if (t == e) {
			return true
		}
	}
	return false
}

func (r *Device) remote_broadcaster() {
	fmt.Println("Starting bcast....")
	fmt.Println(r.profile.IP + ":" + r.profile.TPort)
	target_addr,err := net.ResolveUDPAddr("udp", r.profile.IP + ":" + r.profile.TPort)
	Check(err)
	out_connection, err := net.DialUDP("udp", nil, target_addr)
	Check(err)
	defer out_connection.Close()
	
	fmt.Println("Started bcast!")
	for {
		select {
		case msg := <- r.send:
			encoded, err := json.Marshal(msg)
			Check(err)
			out_connection.Write(encoded)
			fmt.Println("Sent:", msg, "to", r.profile.IP + ":" + r.profile.TPort)
		}
	}
}

func (r *Device) remote_listener(localip string) {
	listen_addr, err := net.ResolveUDPAddr("udp", localip + ":" + r.profile.RPort)
	Check(err)
	in_connection, err := net.ListenUDP("udp", listen_addr)
	Check(err)
	defer in_connection.Close()
	//var ack ack
	
	var message capsule
	
	wd_kick := make(chan bool, 100)
	for {
		fmt.Println("Starts listening on:", r.profile.RPort)
		buffer := make([]byte, 1024)
		length, _, _ := in_connection.ReadFromUDP(buffer)
		if (r.alive == false) {
			go r.watchdog(wd_kick)
			fmt.Println("Connection with device", r.id, "established!")
			r.Connected <- true
		}
		wd_kick <- true
		
		err := json.Unmarshal(buffer[:length], &message)
		Check(err)
		fmt.Println("Received:", message.ItemData)
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
		case PAIRING:
			fmt.Println("Found pairing data, sending to Receive")
			r.Receive <- message.ItemData
		default:
			fmt.Println("Received data:", message.ItemData)
		}
		
	}
}

func (r *Device) handle_ack(d data) {
	//r.tag_rm <- tag(d)
	r.ackchan <- tag(d)
}
func (r *Device) ping_remote() {
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

func (r *Device) send_ack(reference capsule) {
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

func (r *Device) watchdog(kick <- chan bool) {
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
	fmt.Println("Connection with device", r.id, "lost.")
}

func getLocalip() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	Check(err)
	defer conn.Close()

	address := conn.LocalAddr().String()
	ip := stripPort(address)
	
	return ip
}

func stripPort(address string) string {
	var ip string = ""
	for _, char := range address {
		if (char == ':') {
			break
		}
		ip += string(char)
	}
	return ip
}

func (r *Remote) GetSize() int {
	n := r.store.GetSize()
	m := len(r.device)
	Assert(n == m, "Mismatched size of active devices and stored addresses")
	return n
}
