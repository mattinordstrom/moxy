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
	fmt.Println(bullet + " Admin UI:      http://localhost:9097/moxyadminui")
	fmt.Println(" ")
	fmt.Println(" ")

	// Mockdef
	mockObjArr := ut.GetJSONObj("mockdef.json")
	fmt.Println(ut.ColorPurple + "-------- Mockdef --------" + ut.ColorReset)
	for _, val := range mockObjArr {
		val, ok := val.(map[string]interface{})
		if !ok {
			log.Fatalln("mockdef. expected type map[string]interface{} (init)")
		}

		payloadStr := fmt.Sprint(val["payloadFromFile"])
		if !ut.UsePayloadFromFile(val) {
			payloadM, err := json.Marshal(val["payload"])
			if err != nil {
				log.Fatalf("Error occurred during marshalling in main: %v", err)
			}

			payloadStr = string(payloadM)
		}

		fmt.Println(ut.GetMockEventString(val, true, payloadStr) + getInactiveStr(fmt.Sprint(val["active"])))
	}
	fmt.Println(ut.ColorPurple + "-------------------------\n" + ut.ColorReset)

	// Proxydef
	objArr := ut.GetJSONObj("proxydef.json")
	fmt.Println(ut.ColorGreen + "-------- Proxydef --------" + ut.ColorReset)
	for _, val := range objArr {
		val, ok := val.(map[string]interface{})
		if !ok {
			log.Fatalln("proxydef. expected type map[string]interface{} (init)")
		}

		fmt.Println(ut.ColorGreen +
			fmt.Sprint(val["urlpart"]) + ut.RightArrow + fmt.Sprint(val["target"]) +
			ut.ColorReset +
			getInactiveStr(fmt.Sprint(val["active"])))
	}
	fmt.Println(ut.ColorGreen + "--------------------------\n" + ut.ColorReset)

	if !*lFlag {
		// INIT HTTP HANDLING
		hh.CreateHTTPListener()
	}
}

func getInactiveStr(a string) string {
	activeStr := ""
	active, parseError := strconv.ParseBool(a)
	if parseError != nil {
		log.Fatalln("parseError active flag")
	} else if !active {
		activeStr = ut.ColorRed + " [INACTIVE]" + ut.ColorReset
	}

	return activeStr
}
