package main

import (
	"strings"
	"testing"

	testhelper "github.com/mattinordstrom/moxy/testhelpers"
	"github.com/mattinordstrom/moxy/utils"
)

func TestGetEventString(t *testing.T) {
	t.Parallel()

	utils.MockFile = "mockdef_test.json"
	utils.ProxyFile = "proxydef_test.json"

	output := testhelper.CaptureOutput(t, PrintInitSetup)

	// Test mock - print payload
	expected := utils.ColorPurple +
		"GET /api/test/mocktesting  ➔  [0 ms] 404 {\"response\":\"mocktest\"}" +
		utils.ColorReset
	test := strings.Contains(output, expected)

	if !test {
		testhelper.PrintAssertError(t, expected, output)
	}

	// Test mock - print payloadFromFile and freeze time
	expected = utils.ColorPurple +
		"PUT /api/test/mocktesting2  ➔  [1500 ms] 200 /absolute/path/to/some/file" +
		utils.ColorReset
	test = strings.Contains(output, expected)

	if !test {
		testhelper.PrintAssertError(t, expected, output)
	}

	// Test proxy - should be green color
	expected = utils.ColorGreen + "/api/test123  ➔  http://localhost:8088" + utils.ColorReset
	test = strings.Contains(output, expected)

	if !test {
		testhelper.PrintAssertError(t, expected, output)
	}

	// Test proxy - INACTIVE
	expected = utils.ColorGreen + "/api/test456  ➔  http://localhost:8089" + utils.ColorReset +
		utils.ColorRed + " [INACTIVE]" + utils.ColorReset
	test = strings.Contains(output, expected)

	if !test {
		testhelper.PrintAssertError(t, expected, output)
	}
}
