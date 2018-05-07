package addresses
import(
	"fmt"
	"strconv"
	"math/rand"
	"time"
	fs "./filesys"
)
type AddrList struct {
	data		[]AddrData
	file		fs.List
	pairPort	string
}

type AddrData struct {
	IP		string
	TPort		string
	RPort		string
}

func (a *AddrList) Init(path string) {
	var filesys fs.List
	a.file = filesys
	a.file.Init(path)
	a.pairPort = "5555"
	
	capsule := a.file.Load()
	a.data = listParser(capsule)
}

func (a *AddrList) GetData(index int) AddrData {
	data := a.data[index]
	return data
}

func (a *AddrList) GetSize() int {
	n := len(a.data)
	return n
}

func (a *AddrList) Add(data AddrData) {
	if (data.RPort == "") {
		data.RPort = a.makePort()
		fmt.Println("Found new port:",data.RPort)
	}
	a.addAddress(data)
	a.save()
}

func (a *AddrList) Remove(remove AddrData) {
	n := len(a.data) // store length in struct ?
	tmpList := make([]AddrData, n)
	
	if (n > 0) {
		found := 0
		for i := 0; i < n; i++ {
			if (a.data[i] != remove) {
				tmpList[i - found] = a.data[i]
			} else {
				found ++
			}
		}
		if (found > 0) {
			fmt.Println("Found:", found)
			a.data = make([]AddrData, n - found)
			copy(a.data, tmpList)
			fmt.Println("Removed", remove)
			a.save()
		} else {
			fmt.Println("Couldn't find the requested remove.")
		}
	} else {
		fmt.Println("Found nothing to remove.")
	}
}

func (a *AddrList) addAddress(data AddrData) {
	n := len(a.data)
	
	if (n < 1) {
		a.data = make([]AddrData, 1)
		a.data[0] = data
	} else {
		n := len(a.data)
		var add AddrData
		add = data
		tmpList := make([]AddrData, n+1)
		copy(tmpList, a.data)
		a.data = tmpList
		a.data = a.data[0 : n+1]
		a.data[n] = add
	}
}

func (a *AddrList) GetPairPort() string {
	port := portConv(a.pairPort)
	return port
}

func (a *AddrList) makePort() string {
	const MAX_PORT_SIZE int = 65535
	const MIN_PORT_SIZE int = 1023
	GLOBAL_PORT, _  := strconv.Atoi(a.pairPort)
	content := a.data
	n := len(content)
	aContent := make([]AddrData, n)
	portList := make([]int, n)
	for i := 0; i < n; i++ {
		portList[i], _ = strconv.Atoi(aContent[i].RPort)
		//assert(err == nil, "Cannot convert port list to int in func makePort()")
	}
	var port_unique bool
	var counter int
	var assertC int
	var port int
	for {
		counter ++
		for {
			port = randomPort(MAX_PORT_SIZE - MIN_PORT_SIZE)
			port += MIN_PORT_SIZE
			if (port != GLOBAL_PORT) {
				break
			}
			assertC ++
			if (assertC > 10000) {
				fmt.Println("Something strange is happening.. Check makePort() in package addresses!")
			}
		}
		port_unique = true
		for i := 0; i < n; i++ {
			var check int = portList[i]
			if (check == port) {
				port_unique = false
			}
		}
		if (port_unique == true) {
			break
		}
		if (counter > 10000) {
			fmt.Println("func makePort() is hanging. AddrList size:", n)
			counter = 0
		}
	}
	return strconv.Itoa(port)
}

func randomPort(max_value int) int {
	seed := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(seed)
	random := rand.Intn(max_value)
	var r int = int(random)
	return r
}

func (a *AddrList) GetRPort(index int) string {
	port := a.data[index].RPort
	str := portConv(port)
	return str
}

func (a *AddrList) GetTPort(index int) string {
	port := a.data[index].TPort
	str := portConv(port)
	return str
}

func (a *AddrList) GetIP(index int) string {
	IP := a.data[index].IP
	return IP
}


func (a *AddrList) SetTPort(index int, port int) {
	assert(port <= 65535, "Can't set the port bigger than 65535.")
	assert(port >= 1023, "Can't set the port smaller than 1023.")
	a.data[index].TPort = strconv.Itoa(port)
	a.save()
}

func portConv(number string) string {
	inumber, _ := strconv.Atoi(number)
	assert(inumber <= 65535, "A port is bigger than 65535.")
	assert(inumber >= 1023, "A port is smaller than 1023.")
	
	r := []rune(number)
	n := len(number)
	port := make([]rune, (n + 1))
	port[0] = ':'
	for i := 0; i < n; i++ {
		port[i + 1] = r[i]
	}
	return string(port)
}

func listParser(input map[string]fs.Envelop) []AddrData {
	n := fs.GetLength(input)
	out := make([]AddrData, n)
	for i := 0; i < n; i++ {
		ss := input[strconv.Itoa(i)]
		out[i].IP = assertStr(ss["IP"])
		out[i].TPort = assertStr(ss["TPort"])
		out[i].RPort = assertStr(ss["RPort"])
	}
	return out
}

func (a *AddrList) save() {
	n := len(a.data)
	list := make([]interface{}, n)
	for i := 0; i < n; i++ {
		list[i] = a.data[i]
	}
	a.file.Save(list)
}

func assertStr(input interface{}) string {
	var empty string
	if asserted, ok := input.(string); ok {
		return asserted
	} else {
		fmt.Println("Couldn't assert string...")
		return empty
	}
}

func assert(expression bool, message string) {
	if ( !(expression) ) {
		fmt.Println("Assertion FAIL: ", message)
		panic(nil)
	}
}

func (a *AddrList) ClearAllData() {
	a.data = nil
	a.file.Save(nil)
}