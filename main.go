package main
import nw "./network"
import cf "./config"
import ("fmt"
)


const REMOTES = 2
func main() {
	var remote [cf.MAX_REMOTES]nw.Remote

	nw.Init(&remote)
	
	fmt.Println("started...")
	for {
	}
}
