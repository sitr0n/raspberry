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
	
	fmt.Println("started...")
	for {
		time.Sleep(time.Second)
		remote[0].Send(4434)
	}
}
