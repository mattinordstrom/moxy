package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mattinordstrom/moxy/config"
	hh "github.com/mattinordstrom/moxy/httphandling"
	ut "github.com/mattinordstrom/moxy/utils"
)

var bullet = "\u2022"

func main() {
	pFlag := flag.Int("p", hh.Port, "Specify port to run on")
	lFlag := flag.Bool("l", false, "List redirects without starting server")
	sFlag := flag.Bool("s", false, "Run on https")

	flag.Parse()

	hh.Port = *pFlag

	firstTimeSetup()

	err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	hh.DefaultRoute = config.AppConfig.Defaults.DefaultRoute

	protocol := "http"
	if *sFlag {
		protocol = "https"
	}

	fmt.Println(ut.ColorBold + "::::::::::::::: MOXY :::::::::::::::" + ut.ColorReset)
	fmt.Println(ut.ColorGray + "mocking and proxying requests on localhost" + ut.ColorReset)
	fmt.Println(" ")
	fmt.Println(bullet + " Run on:        " + protocol + "://localhost:" + strconv.Itoa(hh.Port))
	fmt.Println(bullet + " WS mock        " + "ws://localhost:" + strconv.Itoa(hh.Port) + "/moxywsmock")
	fmt.Println(bullet + " Default route: " + hh.DefaultRoute)
	fmt.Println(bullet + " Admin UI:      " + protocol + "://localhost:" + strconv.Itoa(hh.Port) + "/moxyadminui")
	fmt.Println(" ")
	fmt.Println(" ")

	PrintInitSetup()

	if !*lFlag {
		hh.CreateHTTPListener(*sFlag)
	}
}

func PrintInitSetup() {
	// Mockdef
	mockObjArr := ut.GetMockJSON()
	fmt.Println(ut.ColorPurple + "-------- Mockdef --------" + ut.ColorReset)

	for _, mockEntity := range mockObjArr {
		payloadStr := mockEntity.PayloadFromFile

		if !ut.UsePayloadFromFile(mockEntity) {
			jsonPayload, err := json.Marshal(mockEntity.Payload)
			if err != nil {
				log.Fatalf("Error occurred during marshalling in main: %v", err)
			}

			payloadStr = string(jsonPayload)
		}

		fmt.Println(ut.GetMockEventString(mockEntity, true, payloadStr) + getInactiveStr(mockEntity.Active))
	}

	fmt.Println(ut.ColorPurple + "-------------------------\n" + ut.ColorReset)

	// Proxydef
	proxyObjArr := ut.GetProxyJSON()
	fmt.Println(ut.ColorGreen + "-------- Proxydef --------" + ut.ColorReset)

	for _, proxyEntity := range proxyObjArr {
		fmt.Println(ut.GetProxyEventString(proxyEntity.URLPart, proxyEntity.Target, "", false) +
			getInactiveStr(proxyEntity.Active))
	}

	fmt.Println(ut.ColorGreen + "--------------------------\n" + ut.ColorReset)
}

func getInactiveStr(active bool) string {
	activeStr := ""
	if !active {
		activeStr = ut.ColorRed + " [INACTIVE]" + ut.ColorReset
	}

	return activeStr
}

func firstTimeSetup() {
	targetExisted, err := ut.CopyFile("templates/config_template.yml", "config.yml")
	if err != nil {
		fmt.Println("File copy failed:", err)
	} else if !targetExisted {
		replaceAbsPaths("config.yml")
	}

	targetExisted, err = ut.CopyFile("templates/mockdef_template.json", "mockdef.json")
	if err != nil {
		fmt.Println("File copy failed:", err)
	} else if !targetExisted {
		replaceAbsPaths("mockdef.json")
	}

	_, err = ut.CopyFile("templates/proxydef_template.json", "proxydef.json")
	if err != nil {
		fmt.Println("File copy failed:", err)
	}
}

func replaceAbsPaths(filename string) {
	_, mainFileFullPath, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get caller info")
	}
	dir := filepath.Dir(mainFileFullPath)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	placeholder := "/absolute_path_to_proj_dir/moxy"

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	updated := strings.ReplaceAll(string(data), placeholder, absDir)

	const filePermission = 0o664
	errr := os.WriteFile(filename, []byte(updated), filePermission)
	if errr != nil {
		log.Fatal(errr)
	}
}
