package main

import (
	"fmt"
	_ "fmt"
	"os"
	_ "os"
	"regexp"
)

var idt string = "(?:[_qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM][_\\w]*)"
var typet string = `[^\(\)]+` //every char except () to prevent consuming more than one set of parens
var paramt string = fmt.Sprintf(`(?:\s*%s\s+%s\s*)`, idt, typet)

var actupong string = fmt.Sprintf(`\s*(\s*\(\s*(?:\s*%s\s+)?\s*\*?\s*%s\s*\)\s*)?\s*`, idt, idt)                                //"(\((?:${idt}\s+)?\*?${idt}\))?" || "\s*(\s*\(\s*(?:\s*${idt}\s+)?\s*\*?\s*${idt}\s*\)\s*)?\s*"
var paramlistg string = fmt.Sprintf(`\s*(\s*\(\s*(?:\s*%s\s*(?:\s*\,\s*%s\s*)*\s*)?\s*\)\s*)\s*`, paramt, paramt)               //"(\((?:${paramt}(?:\,${paramt})*)?\))" || "\s*(\s*\(\s*(?:\s*${paramt}\s*(?:\s*\,\s*${paramt}\s*)*\s*)?\s*\)\s*)\s*"
var returnlistg string = fmt.Sprintf(`\s*((?:\s*\(\s*(?:\s*%s\s*(?:\s*\,\s*%s\s*)*\s*)?\s*\)\s*)|%s)\s*?`, typet, typet, typet) //"((?:\((?:${typet}(?:\,${typet})*)?\))|${typet})?" || "\s*((?:\s*\(\s*(?:\s*${typet}\s*(?:\s*\,\s*${typet}\s*)*\s*)?\s*\)\s*)|${typet})\s*?"
var signatureg string = fmt.Sprintf(`\s*func\s*%s\s*(%s)\s*%s\s*%s\s*\{\s*`, actupong, idt, paramlistg, returnlistg)            //"(%{idt})(\(%{paramlistt}\))(\(%{paramlistt}\))"

var sigReg *regexp.Regexp
var sigs []*Signature
var notSigs []string

func init() {
	sigReg = regexp.MustCompile(signatureg)
}

func main() {
	fmt.Println(signatureg)

	ProcessFile("tx.pb.go.dat")

	fmt.Println("Processed signatures:", len(sigs))
	fmt.Println("Uncaught sigs:", len(notSigs))
	for _, a := range notSigs {
		fmt.Println(a)
	}

	WriteOutData("signature_list.txt")
}

func WriteOutData(filestr string) {

	file, err := os.Create(filestr)
	if err != nil {
		panic(err)
	}

	for _, a := range sigs {
		file.Write([]byte(a.ToString()))
	}

	defer file.Close()
}

func ProcessFile(fileStr string) {

	file, err := os.Open(fileStr)
	if err != nil {
		panic(err)
	}

	stat, statErr := file.Stat()
	if statErr != nil {
		panic(statErr)
	}

	data := make([]byte, stat.Size())
	numBytesRead, readErr := file.Read(data)
	if readErr != nil {
		panic(readErr)
	} else if numBytesRead != int(stat.Size()) {
		fmt.Println("ERROR: Number of bytes read != size of file")
	}

	reader := NewReader(data)

	counter := 0
	for !reader.Eof() {
		line := reader.Readln()

		if Contains([]byte("func"), line) {

			ProcessLine(string(line))
			counter++
		}
	}

	fmt.Println("Func lines found:", counter)

	defer file.Close()
}

func ProcessLine(input string) {

	regex := sigReg.FindStringSubmatch(input)
	if len(regex) >= 1 {
		sigs = append(sigs, SigFromRegex(regex))
	} else {
		notSigs = append(notSigs, input)
	}
}

func Contains(target []byte, repo []byte) bool {

	if len(target) > len(repo) {
		return false
	}

	for ir, cr := range repo {

		if cr == target[0] {

			ret := true
			for it := range target {

				if !(ir+it < len(repo) && target[it] == repo[ir+it]) {
					ret = false
				}
			}

			if ret {
				return ret
			}
		}
	}

	return false
}

func testContains() {
	a := []byte("Test 123")
	b := []byte("Test")
	fmt.Println(true, Contains(b, a))

	a = []byte("Test 123")
	b = []byte("123")
	fmt.Println(true, Contains(b, a))

	a = []byte("Test 123")
	b = []byte("Test 123")
	fmt.Println(true, Contains(b, a))

	a = []byte("Test 123")
	b = []byte("Testr")
	fmt.Println(false, Contains(b, a))
}

//Used for debugging regex
func PrintRegexTypes() {
	fmt.Println("Identifier:---------------------")
	fmt.Println(idt)
	fmt.Println("Param List:---------------------")
	fmt.Println(paramlistg)
	fmt.Println("Return List:---------------------")
	fmt.Println(returnlistg)
	fmt.Println("Signature:---------------------")
	fmt.Println(signatureg)
}

//Used for debugging regex
func PrintRegexGroups() {
	fmt.Println("Act Upon:---------------------")
	fmt.Println(actupong)
	fmt.Println("Identifier:---------------------")
	fmt.Println(idt)
	fmt.Println("Identifier:---------------------")
	fmt.Println(idt)
}

/* ============= Signature Object =============== */

type Signature struct {
	actupon    string
	id         string
	params     string
	returnData string

	raw string
}

func SigFromRegex(regex []string) *Signature {
	return &Signature{
		actupon:    regex[1],
		id:         regex[2],
		params:     regex[3],
		returnData: regex[4],

		raw: regex[0],
	}
}

func (s *Signature) ToString() string {
	return fmt.Sprintf("----- %s -----\nActs upon:\t'%s'\nParameters:\t'%s'\nReturns:\t'%s'\nRaw line:\t'%s'\n\n", s.id, s.actupon, s.params, s.returnData, s.raw)
}
