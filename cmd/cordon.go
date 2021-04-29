package cmd

import (
        "fmt"
        "github.com/spf13/cobra"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getNodesInAZ(zone string) []string {
        k8sNodes := make([]string, 0)
        nodes, err := Clientset.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: "failure-domain.beta.kubernetes.io/zone="+zone})
        if err != nil {
                fmt.Println("error listing nodes in az: %v", err)
        }
        for node, _ := range nodes.Items {
                k8sNodes = append(k8sNodes, nodes.Items[node].ObjectMeta.Labels["kubernetes.io/hostname"])
        }
        return k8sNodes
}

var cordonAZ = &cobra.Command{
        Use: "cordonAZ",
        Short: "Cordon an AZ (only one)",
        Long: "This will make sure the nodes are unschedulable, so when the drain command runs nodes won't go back to the bad az",
        Run: func (cmd *cobra.Command, args []string) {
                fmt.Println("Need to impliment scaling cluster auto scaler down as part of this")
                fmt.Println("Cordon!"+ zone)
                fmt.Println(getNodesInAZ(zone))
        },
}
