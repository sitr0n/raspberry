package main

import ("fmt"
)
import nw "./network"

const REMOTES = 1
func main() {
	var remote [REMOTES]nw.Remote
	var addr []string
	//addr[0] = "10"
	//addr[1] = 11
	nw.Init(addr, &remote)
	
	fmt.Println("started...")
	for {
	}
}
