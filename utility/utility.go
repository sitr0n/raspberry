package utility
import(
	"fmt"
	"time"
)

func Assert(expression bool, message string) {
	if ( !(expression) ) {
		fmt.Println("Assertion FAIL: ", message)
		panic(nil)
	}
}

func Timer(length int, stop <- chan bool, timeout chan <- bool) {
	interrupted := false
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(length)*time.Millisecond)
		select {
		case <- stop:
			interrupted = true
		default:
		}
	}
	if (!interrupted) {
		timeout <- true
	}
}

func AssertInt64(d interface{}) int {
	if asserted, ok := d.(float64); ok {
		return int(asserted)
	} else {
		fmt.Println("Couldn't assert int.")
		return 0
	}
}

func AssertInt(d interface{}) int {
	if asserted, ok := d.(int); ok {
		return asserted
	} else {
		fmt.Println("Couldn't assert int.")
		return 0
	}
}

func Check(e error) {
	if (e != nil) {
		panic(e)
	}
}