package cmd

import (
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "fmt"
        "os"
        "io"
        "github.com/spf13/cobra"
)


func GetFailureDomains() map[string]int {
        availability_zones := make(map[string]int)
        nodes, err := Clientset.CoreV1().Nodes().List(metav1.ListOptions{})
        if err != nil {
                fmt.Printf("error listing nodes: %v", err)
        }
        for node, _ := range nodes.Items {
                availability_zones[nodes.Items[node].ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"]] += 1
        }
        return availability_zones
}

func DisplayAvailabilityZones(writer io.Writer, zoneMap map[string]int) {
        message := ""
        for key, value := range zoneMap{
                message = message + fmt.Sprintf("AZ: %s\tNumber of nodes: %d\n", key, value)
        }
        fmt.Fprintf(writer, message)
}


var listAZs = &cobra.Command{
        Use: "listAZs",
        Short: "List AZs",
        Long: "Lists all the availability zones associated to your current context",
        Run: func(cmd *cobra.Command, args []string) {
                availability_zones := getFailureDomains()
                DisplayAvailabilityZones(os.Stdout, availability_zones)
        },
}
