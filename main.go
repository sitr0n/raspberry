package main
import nw "./network"
import _ "./config"
import ("fmt"
	_ "time"
)


const REMOTES = 2
func main() {
	test := make(chan bool)
	var remote nw.RemoteList
	fmt.Println("Initing..")
	remote.Init()
	fmt.Println("done..")
	fmt.Println(remote.GetSize())
	<- test
	//var r nw.Remote
	
	//nw.Add_remote("02", "20", &r)
	//fmt.Println(*r.id)
	//go rec(r)
	/*
	go rec(&remote[0])
	fmt.Println("started...")
	for {
		time.Sleep(2*time.Second)
		remote[0].Send(244)
	}
	*/
}

func rec(remote *nw.Remote) {
	for {
		select {
		case input := <- remote.Receive:
			fmt.Println(input)
		}
	}
}
