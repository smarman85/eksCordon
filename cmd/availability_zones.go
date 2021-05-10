package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
)

func GetFailureDomains(clientset kubernetes.Interface) (map[string]int, error) {
	availability_zones := make(map[string]int)
	nodes, err := clientset.
		CoreV1().
		Nodes().
		List(metav1.ListOptions{})
	if err != nil {
		//fmt.Printf("error listing nodes: %v", err)
		return nil, err
	}
	for node, _ := range nodes.Items {
		availability_zones[nodes.Items[node].ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"]] += 1
	}
	return availability_zones, nil
}

func DisplayAvailabilityZones(writer io.Writer, zoneMap map[string]int) {
	message := ""
	for key, value := range zoneMap {
		message = message + fmt.Sprintf("AZ: %s\tNumber of nodes: %d\n", key, value)
	}
	fmt.Fprintf(writer, message)
}

var listAZs = &cobra.Command{
	Use:   "listAZs",
	Short: "List AZs",
	Long:  "Lists all the availability zones associated to your current context",
	Run: func(cmd *cobra.Command, args []string) {
		availability_zones, err := GetFailureDomains(client)
		if err != nil {
			fmt.Printf("error listing nodes: %v", err)
		}
		DisplayAvailabilityZones(os.Stdout, availability_zones)
	},
}
