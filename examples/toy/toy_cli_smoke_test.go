package toy_test

import (
	"testing"

	toyvcr "github.com/xeger/goa-vcr/examples/toy/gen/http/toy/vcr"
	toy "github.com/xeger/goa-vcr/examples/toy/gen/toy"
)

func TestVCRCLI_Usage(t *testing.T) {
	code := toyvcr.RunCLI([]string{"help"}, toyvcr.CLIConfig{
		AppName: "toy-vcr",
		ScenarioRegistry: map[string]func(toy.Service) toy.Service{
			"Noop": func(next toy.Service) toy.Service {
				return next
			},
		},
		DefaultPort:      8080,
		DefaultUpstream:  "https://example.com",
		DefaultScenarios: []string{"Noop"},
	})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}
