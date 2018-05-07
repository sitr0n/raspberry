package utility
import(
	"fmt"
)

func Assert(expression bool, message string) {
	if ( !(expression) ) {
		fmt.Println("Assertion FAIL: ", message)
		panic(nil)
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