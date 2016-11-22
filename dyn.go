package main


import (
	"flag"
	"fmt"
	"os"
	"bufio"
	"strings"
	"syscall"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"golang.org/x/crypto/ssh/terminal"
)

const VERSION = "0.0.1"
const BASEURL = "https://api.dynect.net/REST/"


func check_arg (arg string, validArgs []string) bool {
	var valid bool
	for _, val := range validArgs {
		if arg == val {
			valid = true
			break
		} else {
			valid = false
		}
	}
	return valid
}

/*
func freak_the_fuck_out (e error) {
	if e != nil {
		panic(e)
	}
}
*/


func type_of(v interface{}) string {
    return fmt.Sprintf("%T", v)
}


func get_credentials_from_file (fileName string) map[string]string {
	var authData = make(map[string]string)
	inputFile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, "=")
		authData[split[0]] = split[1]
	}
	return authData
}


func get_credentials_interactively () map[string]string {
	var authData = make(map[string]string)
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your Dyn customer name: ")
	customerName, _ := reader.ReadString('\n')
	authData["customer_name"] = strings.TrimSpace(customerName)
	
	fmt.Print("Enter your Dyn username: ")
	userName, _ := reader.ReadString('\n')
	authData["user_name"] = strings.TrimSpace(userName)

	fmt.Print("Enter your Dyn password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		fmt.Println("\nPassword typed: " + string(bytePassword))
	}

	password := strings.TrimSpace(string(bytePassword))
	authData["password"] = password
	return authData
}


type DynSession struct {
	BaseUrl string
	Token   string
}


func (d *DynSession) login (customerName, userName, password string) {
	apiEndPoint := "Session/"
	type AuthStruct struct {
		Customer_Name string `json:"customer_name"`
		User_Name     string `json:"user_name"`
		Password      string `json:"password"`
	}
	auth := AuthStruct{
		Customer_Name: customerName,
	        User_Name:     userName,
	        Password:      password,
	}
	authJson, err := json.Marshal(auth)
	if err != nil {
		fmt.Println("error: ", err)
	}
	
	req, err := http.Post(d.BaseUrl + apiEndPoint,
		"application/json",
		bytes.NewBuffer(authJson))
	if err != nil {
		fmt.Println("error: ", err)
	}

	defer req.Body.Close()

	res, _ := ioutil.ReadAll(req.Body)
	
	var resMap map[string]interface{}
	if err := json.Unmarshal(res, &resMap); err != nil {
		fmt.Println("cannot parse JSON")
		panic(err)
	}
	data := resMap["data"].(map[string]interface{})
	d.Token = data["token"].(string)
}


func (d *DynSession) logout () {
	apiEndpoint := "Session/"
	client := http.Client{}
	req, err := http.NewRequest("DELETE", d.BaseUrl + apiEndpoint, nil)
	req.Header.Add("Auth-Token", d.Token)
	fmt.Println(err)

	resp, err := client.Do(req)

	defer resp.Body.Close()
	
	fmt.Println(resp)
}


func main () {
	versionArg := flag.Bool("version", false, "print version string")
	serviceArg := flag.String("service", "", "service/fqdn to check")
	modeArg := flag.String("mode", "", "mode to set a service {enable,disable}")
	validModes := []string{"enable", "disable"}
//	statusArg := flag.Bool("status", true, "get the current state of a service")
	fileArg := flag.String("file", "", "file containt credentials for Dyn")
//	ipArg := flag.String("ip", "", "ip address to manage")
	customerArg := flag.String("customer_name", "", "Dyn customer name")
	userArg := flag.String("user_name", "", "Dyn username for logging in")
	passwordArg := flag.String("password", "", "Dyn password for logging in")
	flag.Parse()
	authData := make(map[string]string)

	// print version
	if *versionArg {
		fmt.Printf("dyn version %s\n", VERSION)
		os.Exit(0)
	}

	// required argument
	if *serviceArg == "" {
		fmt.Println("--service argument is required")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// choice argument
	if *modeArg != "" && ! check_arg(*modeArg, validModes) {
		fmt.Printf("Arg: '%s' not recognized" , *modeArg)
		fmt.Printf("valid arguments are %s", validModes)
		os.Exit(1)
	}

	if *fileArg != "" {
		authData = get_credentials_from_file(*fileArg)
	} else if *customerArg != "" && *userArg != "" && *passwordArg != "" {
		authData["customer_name"] = *customerArg
		authData["user_name"] = *userArg
		authData["password"] = *passwordArg
	} else {
		authData = get_credentials_interactively()
	}

	session := DynSession{
		BaseUrl: BASEURL,
	}
	session.login(authData["customer_name"], authData["user_name"], authData["password"])
	// session.logout()
	os.Exit(0)
}
