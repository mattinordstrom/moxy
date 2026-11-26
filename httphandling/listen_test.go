package httphandling_test

import (
	"slices"
	"testing"

	"github.com/mattinordstrom/moxy/httphandling"
)

//nolint:paralleltest // reads the global port
func TestListenAddrs(t *testing.T) {
	orig := httphandling.Port
	httphandling.Port = 9097

	t.Cleanup(func() { httphandling.Port = orig })

	if got, want := httphandling.ListenAddrs(false), []string{ipv4Host, ipv6Host}; !slices.Equal(got, want) {
		t.Errorf("localhost only: got %v, want %v", got, want)
	}

	if got, want := httphandling.ListenAddrs(true), []string{":9097"}; !slices.Equal(got, want) {
		t.Errorf("all interfaces: got %v, want %v", got, want)
	}
}
