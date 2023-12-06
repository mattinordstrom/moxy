package main

import (
	"strings"
	"testing"

	testhelper "github.com/mattinordstrom/moxy/testhelpers"
	"github.com/mattinordstrom/moxy/utils"
)

func TestMain(m *testing.M) {
	testhelper.SetupTest(m)
}

func TestGetEventString(t *testing.T) {
	t.Parallel()

	output := testhelper.CaptureOutput(t, PrintInitSetup)

	// Test mock - print payload
	expected := utils.ColorPurple +
		"GET /api/test/mocktesting  ➔  [0 ms] 404 {\"response\":\"mocktest\"}" +
		utils.ColorReset
	test := strings.Contains(output, expected)

	if !test {
		testhelper.PrintAssertError(t, expected, output, "")
	}

	// Test mock - print payloadFromFile and freeze time
	expected = utils.ColorPurple +
		"PUT /api/test/mocktesting2  ➔  [1500 ms] 200 /absolute/path/to/some/file" +
		utils.ColorReset
	test = strings.Contains(output, expected)

	if !test {
		testhelper.PrintAssertError(t, expected, output, "")
	}

	// Test proxy - should be green color
	expected = utils.ColorGreen + "/api/test123  ➔  http://localhost:8088/api/test123" + utils.ColorReset
	test = strings.Contains(output, expected)

	if !test {
		testhelper.PrintAssertError(t, expected, output, "")
	}

	// Test proxy - INACTIVE
	expected = utils.ColorGreen + "/api/test456  ➔  http://localhost:8089/api/test456" + utils.ColorReset +
		utils.ColorRed + " [INACTIVE]" + utils.ColorReset
	test = strings.Contains(output, expected)

	if !test {
		testhelper.PrintAssertError(t, expected, output, "")
	}
}
