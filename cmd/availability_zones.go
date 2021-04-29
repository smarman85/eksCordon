package cmd

import (
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "fmt"
        "github.com/spf13/cobra"
)


func getFailureDomains() map[string]int {
        availability_zones := make(map[string]int)
        nodes, err := Clientset.CoreV1().Nodes().List(metav1.ListOptions{})
        if err != nil {
                fmt.Println("error listing nodes: %v", err)
        }
        for node, _ := range nodes.Items {
                availability_zones[nodes.Items[node].ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"]] += 1
        }
        return availability_zones
}

func displayAvailabilityZones(zoneMap map[string]int) {
        for key, value := range zoneMap{
                fmt.Println("AZ: ", key, "\tNumber of nodes: ", value)
        }
}


var listAZs = &cobra.Command{
        Use: "listAZs",
        Short: "List AZs",
        Long: "Lists all the availability zones associated to your current context",
        Run: func(cmd *cobra.Command, args []string) {
                availability_zones := getFailureDomains()
                displayAvailabilityZones(availability_zones)
        },
}
