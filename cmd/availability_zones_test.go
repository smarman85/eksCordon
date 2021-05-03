package cmd

import (
        "testing"
        "bytes"
        //"k8s.io/client-go/kubernetes/fake"
)

func TestDisplayAvailabilityZones(t *testing.T) {
        buffer := bytes.Buffer{}
        zones := map[string]int{
                "us-east-1a": 17,
        }
        DisplayAvailabilityZones(&buffer, zones)

        got := buffer.String()
        want := "AZ: us-east-1a\tNumber of nodes: 17\n"

        if got != want {
                t.Errorf("got %q want %q", got, want)
        }
}
