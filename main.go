package main

import ("fmt"
)
import nw "./network"

const REMOTES = 2
func main() {
	var remote [REMOTES]nw.Remote

	nw.Init(&remote)
	
	fmt.Println("started...")
	for {
	}
}
