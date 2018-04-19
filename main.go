package main
import nw "./network"
import cf "./config"
import ("fmt"
	"time"
)


const REMOTES = 2
func main() {
	var remote [cf.MAX_REMOTES]nw.Remote

	nw.Init(&remote)
	
	go rec(&remote[0])
	fmt.Println("started...")
	for {
		time.Sleep(2*time.Second)
		remote[0].Send(244)
	}
}

func rec(remote *nw.Remote) {
	for {
		select {
		case input := <- remote.Receive:
			fmt.Println(input)
		}
	}
}
