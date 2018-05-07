package network
import(
	"fmt"
	addr "./addresses"
)

func main(){
	var Ele0 = addr.AddrData{"123.123.69",":13337", ":420"}
	var Ele1 = addr.AddrData{"321.321.96",":3600", ""}
	msg := make([]interface{}, 1)
	var a addr.AddrList
	a.Init("./addresses/data.json")

	msg[0] = Ele0
	//msg[1] = Ele1
	
	
	//a.Add(
	prnt(a)
	fmt.Println("adding..\n")
	a.Add(Ele1)
	prnt(a)
	
	fmt.Println("removing..\n")
	a.Remove(a.GetData(1))
	prnt(a)
	
	fmt.Println("Getting data:\n")
	fmt.Println(a.GetData(0).IP)
}


func prnt(d addr.AddrList) {
	n := d.GetSize()
	for i := 0; i < n; i++ {
		fmt.Println(i, ":", d.GetData(i))
	}
}