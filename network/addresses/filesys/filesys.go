package filesys
import(
	"fmt"
	"encoding/json"
	"strconv"
	"io/ioutil"
)
/***********************************
* Interface:
* - Use Init with path to file location
* - Use Save with the slice of structs
* - Parse the slice from envelop map to
*   the original struct after Load
*  (See example.go)
*
* TODO:
* Make a list parser function inside
* this package with struct type as
* a parameter. (Look into reflect)
***********************************/
type Envelop map[string]interface{}
type meta struct {
	Length		string
}
type List struct {
	members		[]interface{}
	path		string
}

func GetLength(list map[string]Envelop) int {
	length := list["n"]["Length"]
	assert(length != nil, "File is missing meta data length.")
	aLength := assertStr(length)
	n, err := strconv.Atoi(aLength)
	assert(err == nil, "Trouble converting list length to int")
	return n
}

func (a *List) Init(path string) {
	a.path = path
	//Check the path, make a file if none found
}

func (a *List) Load() map[string]Envelop {
	encoded, err := ioutil.ReadFile(a.path)
	assert(err == nil, "Couldn't read file.")
	
	var mm map[string]Envelop
	err = json.Unmarshal(encoded, &mm)
	assert(err == nil, "Couldn't unmarshal list.")
	return mm
}

func (a *List) Save(list []interface{}) {
	n := len(list)
	var capsule = make(Envelop)
	capsule["n"] = meta{strconv.Itoa(n)}
	
	for i := 0; i < n; i++ {
		capsule[strconv.Itoa(i)] = list[i]
	}
	
	info, err := json.Marshal(&capsule)
	assert(err == nil, "Couldn't marshal list.")
	
	err = ioutil.WriteFile(a.path, info, 0644)
	assert(err == nil, "Couldn't write to file.")
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