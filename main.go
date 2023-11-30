package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/mattinordstrom/moxy/config"
	hh "github.com/mattinordstrom/moxy/httphandling"
	"github.com/mattinordstrom/moxy/models"
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
		// Create struct
		jsonData, err := json.Marshal(val)
		if err != nil {
			fmt.Println("error marshalling map:", err)

			return
		}
		var mockEntity models.Mock
		err = json.Unmarshal(jsonData, &mockEntity)
		if err != nil {
			fmt.Println("error unmarshalling JSON:", err)

			return
		}
		///////////

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
	objArr := ut.GetJSONObj("proxydef.json")
	fmt.Println(ut.ColorGreen + "-------- Proxydef --------" + ut.ColorReset)
	for _, val := range objArr {
		val, ok := val.(map[string]interface{})
		if !ok {
			log.Fatalln("proxydef. expected type map[string]interface{} (init)")
		}

		active, parseError := strconv.ParseBool(fmt.Sprint(val["active"]))
		if parseError != nil {
			log.Fatalln("parseError active flag")
		}

		fmt.Println(ut.ColorGreen +
			fmt.Sprint(val["urlpart"]) + ut.RightArrow + fmt.Sprint(val["target"]) +
			ut.ColorReset +
			getInactiveStr(active))
	}
	fmt.Println(ut.ColorGreen + "--------------------------\n" + ut.ColorReset)

	if !*lFlag {
		// INIT HTTP HANDLING
		hh.CreateHTTPListener()
	}
}

func getInactiveStr(active bool) string {
	activeStr := ""
	if !active {
		activeStr = ut.ColorRed + " [INACTIVE]" + ut.ColorReset
	}

	return activeStr
}
