package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/mattinordstrom/moxy/config"
	hh "github.com/mattinordstrom/moxy/httphandling"
	ut "github.com/mattinordstrom/moxy/utils"
)

var bullet = "\u2022"

func main() {
	pFlag := flag.Int("p", hh.Port, "specify port to run on")
	lFlag := flag.Bool("l", false, "list redirects without starting server")

	flag.Parse()

	hh.Port = *pFlag

	firstTimeSetup()

	err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	hh.DefaultRoute = config.AppConfig.Defaults.DefaultRoute

	fmt.Println(ut.ColorBold + "::::::::::::::: MOXY :::::::::::::::" + ut.ColorReset)
	fmt.Println(ut.ColorGray + "mocking and proxying requests on localhost" + ut.ColorReset)
	fmt.Println(" ")
	fmt.Println(bullet + " Run on:        http://localhost:" + strconv.Itoa(hh.Port))
	fmt.Println(bullet + " Default route: " + hh.DefaultRoute)
	fmt.Println(bullet + " Admin UI:      http://localhost:" + strconv.Itoa(hh.Port) + "/moxyadminui")
	fmt.Println(" ")
	fmt.Println(" ")

	PrintInitSetup()

	if !*lFlag {
		hh.CreateHTTPListener()
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
	err := ut.CopyFile("config_template.yml", "config.yml")
	if err != nil {
		fmt.Println("File copy failed:", err)
	}

	err = ut.CopyFile("mockdef_template.json", "mockdef.json")
	if err != nil {
		fmt.Println("File copy failed:", err)
	}

	err = ut.CopyFile("proxydef_template.json", "proxydef.json")
	if err != nil {
		fmt.Println("File copy failed:", err)
	}
}
