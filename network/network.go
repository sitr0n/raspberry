package network
import cf "../config"
import (
	"fmt"
	"net"
	"time"
	"encoding/json"
	//"os"
	//"strconv"
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

type Remote struct {
	Receive		chan interface{}
	Connected	chan bool
	id		int
	info	  	remote_info
	port		string
	alive 		bool
	send		chan capsule
	tag_req		chan bool
	tag_rm		chan tag
	tag_grant	chan tag
	ackchan		chan tag
	rm_list		[]tag
}

type remote_info struct {
	Name		string
	IP		string
}

var _localip string
var _REMOTES int
func Init(r *[cf.MAX_REMOTES]Remote) {
	_localip = get_localip()
	address_list := load_address()
	
	for i := 0; i < _REMOTES; i++ {
		r[i].Receive		= make(chan interface{}, 100)
		r[i].Connected 		= make(chan bool, 100)
		r[i].id 		= i
		r[i].info 		= address_list[i]
		r[i].port		= create_port(i)
		r[i].alive 		= false
		r[i].send 		= make(chan capsule, 100)
		r[i].tag_req 		= make(chan bool, 100)
		r[i].tag_rm 		= make(chan tag, 100)
		r[i].tag_grant		= make(chan tag, 100)
		r[i].ackchan 		= make(chan tag, 100)
		r[i].rm_list		= []tag{}
		
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
		if (ok == true || (miss < 3) || !r.alive) {
			//fmt.Println("DataAck chain complete!")
			break
		}
		packet.ItemTag = r.create_tag()
	}
}

func (r *Remote) check_for_ack(t tag) bool {
	for _, e := range r.rm_list {
		if (t == e) {
			return true
		}
	}
	return false
}


func (r *Remote) remote_broadcaster() {
	target_addr,err := net.ResolveUDPAddr("udp", r.info.IP + r.port)
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
	listen_addr, err := net.ResolveUDPAddr("udp", _localip + r.port)
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






func check(e error) {
	if (e != nil) {
		panic(e)
	}
}

func Get_remotes() int {
	return _REMOTES
}
