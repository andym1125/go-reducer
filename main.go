package main

import (
	"fmt"
	_ "fmt"
	"os"
	_ "os"
	"regexp"
)

//Initial match regex
var idt string = "(?:[_qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM][_\\w]*)"
var typet string = `[^\(\)\{\}]+` //every char except () to prevent consuming more than one set of parens
var paramt string = fmt.Sprintf(`(?:\s*%s\s+%s\s*)`, idt, typet)

var actupong string = fmt.Sprintf(`\s*(\s*\(\s*(?:\s*%s\s+)?\s*\*?\s*%s\s*\)\s*)?\s*`, idt, idt)                                //"(\((?:${idt}\s+)?\*?${idt}\))?" || "\s*(\s*\(\s*(?:\s*${idt}\s+)?\s*\*?\s*${idt}\s*\)\s*)?\s*"
var paramlistg string = fmt.Sprintf(`\s*(\s*\(\s*(?:\s*%s\s*(?:\s*\,\s*%s\s*)*\s*)?\s*\)\s*)\s*`, paramt, paramt)               //"(\((?:${paramt}(?:\,${paramt})*)?\))" || "\s*(\s*\(\s*(?:\s*${paramt}\s*(?:\s*\,\s*${paramt}\s*)*\s*)?\s*\)\s*)\s*"
var returnlistg string = fmt.Sprintf(`\s*((?:\s*\(\s*(?:\s*%s\s*(?:\s*\,\s*%s\s*)*\s*)?\s*\)\s*)|%s)\s*?`, typet, typet, typet) //"((?:\((?:${typet}(?:\,${typet})*)?\))|${typet})?" || "\s*((?:\s*\(\s*(?:\s*${typet}\s*(?:\s*\,\s*${typet}\s*)*\s*)?\s*\)\s*)|${typet})\s*?"
var signatureg string = fmt.Sprintf(`\s*func\s*%s\s*(%s)\s*%s\s*%s\s*\{\s*`, actupong, idt, paramlistg, returnlistg)            //"(%{idt})(\(%{paramlistt}\))(\(%{paramlistt}\))"

//Cleanup regex
var cleanupToken *regexp.Regexp = regexp.MustCompile(`[^\(\)\{\}\,\s]+`)

//State vars
var sigReg *regexp.Regexp
var sigs []*Signature
var notSigs []string

func init() {
	sigReg = regexp.MustCompile(signatureg)
}

func main() {
	fmt.Println(signatureg)

	fmt.Println((`()        `))
	fmt.Println((`(b []byte) `))
	fmt.Println((`(b []byte, deterministic bool) `))

	ProcessFile("tx.pb.go.dat")

	fmt.Println("Processed signatures:", len(sigs))
	fmt.Println("Uncaught sigs:", len(notSigs))
	for _, a := range notSigs {
		fmt.Println(a)
	}

	WriteOutSigs("signature_list.txt")
	QuickWrite("ids.txt", GetAllIdRaw())
	QuickWrite("params.txt", GetAllParamsRaw())
	QuickWrite("actupon.txt", GetAllActuponRaw())
	QuickWrite("returns.txt", GetAllReturnDataRaw())
}

/* ========== File Handling ========== */

//For debugging. Writes data to file, separated by newline
func QuickWrite(filestr string, data []string) {

	file, err := os.Create(filestr)
	if err != nil {
		panic(err)
	}

	for _, s := range data {
		file.Write([]byte(s))
		file.Write([]byte("\n"))
	}

	defer file.Close()
}

func WriteOutSigs(filestr string) {

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
	for !reader.Eof() {
		line := reader.Readln()

		if Contains([]byte("func"), line) {
			ProcessLine(string(line))
		}
	}

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

/* ============= Signature Object =============== */

type Signature struct {
	actupon_raw    string
	id_raw         string
	params_raw     string
	returnData_raw string

	raw string
}

func SigFromRegex(regex []string) *Signature {
	return &Signature{
		actupon_raw:    regex[1],
		id_raw:         regex[2],
		params_raw:     regex[3],
		returnData_raw: regex[4],

		raw: regex[0],
	}
}

func (s *Signature) ToString() string {
	return fmt.Sprintf("----- %s -----\nActs upon:\t'%s'\nParameters:\t'%s'\nReturns:\t'%s'\nRaw line:\t'%s'\n\n", s.id_raw, s.actupon_raw, s.params_raw, s.returnData_raw, s.raw)
}

func CleanupParams(input string) []string {

	return cleanupToken.FindAllString(input, -1)
}

//For debugging. Returns list of all actupons for testing
func GetAllActuponRaw() []string {
	var ret []string
	for _, a := range sigs {
		ret = append(ret, "'"+a.actupon_raw+"'")
	}
	return ret
}

//For debugging. Returns list of all actupons for testing
func GetAllIdRaw() []string {
	var ret []string
	for _, a := range sigs {
		ret = append(ret, "'"+a.id_raw+"'")
	}
	return ret
}

//For debugging. Returns list of all actupons for testing
func GetAllParamsRaw() []string {
	var ret []string
	for _, a := range sigs {
		ret = append(ret, "'"+a.params_raw+"'")
	}
	return ret
}

//For debugging. Returns list of all actupons for testing
func GetAllReturnDataRaw() []string {
	var ret []string
	for _, a := range sigs {
		ret = append(ret, "'"+a.returnData_raw+"'")
	}
	return ret
}

/* ========== Misc ========== */

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
