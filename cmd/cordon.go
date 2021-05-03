package cmd

import (
        "fmt"
        "github.com/spf13/cobra"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//        "errors"
//        "log"
)

func toggleClusterAutoScaler(desiredReplicas int) {

        currentScale, err := Clientset.AppsV1().
            Deployments("kube-system").
            GetScale("cluster-autoscaler", metav1.GetOptions{})

        if err != nil {
                fmt.Printf("error getting current cluster-autoscaler replicas: %v", err)
        }

        updatedScale := *currentScale
        updatedScale.Spec.Replicas = int32(desiredReplicas)

        scaledConfig, err := Clientset.AppsV1().
            Deployments("kube-system").
            UpdateScale("cluster-autoscaler", &updatedScale)

        if err != nil {
                fmt.Printf("error scaling cluster-autoscaler: %v", err)
        }
        fmt.Println(scaledConfig)
}

//func getNodesInAZ(zone string) ([]string, error) {
func getNodesInAZ(zone string) []string {
        k8sNodes := make([]string, 0)
        nodes, err := Clientset.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: "failure-domain.beta.kubernetes.io/zone="+zone})
        if err != nil {
                fmt.Printf("error listing nodes in az: %v", err)
        }
        for node, _ := range nodes.Items {
                k8sNodes = append(k8sNodes, nodes.Items[node].ObjectMeta.Labels["kubernetes.io/hostname"])
        }
        if len(k8sNodes) == 0 {
                //return k8sNodes, errors.New("There are no nodes in the given zone")
                k8sNodes = append(k8sNodes, "docker-desktop")
        }
        return k8sNodes
}

func cordonNodes(nodesInAZ []string) {
        patch := []byte(`{"spec":{"unschedulable":true}}`)
        for i := 0; i < len(nodesInAZ); i ++ {
                _, err := Clientset.CoreV1().
                    Nodes().
                    Patch(nodesInAZ[i], "application/strategic-merge-patch+json", patch)
                if err != nil {
                        fmt.Printf("error patching node: %v", err)
                }
                fmt.Printf("Successfully Cordoned node: %v", nodesInAZ[i])
        }
}

func podsOnNode(nodeName string) map[string]string {
        //podsOnNode := make([]string, 0)
        podsOnNode := make(map[string]string, 0)
        pods, err := Clientset.CoreV1().
            Pods("").List(metav1.ListOptions{FieldSelector: "spec.nodeName="+nodeName})
        if err != nil {
                fmt.Printf("error listing pods on node: %v %v", nodeName, err)
        }
        for pod, _ := range pods.Items {
                //fmt.Println(pods.Items[pod].Name, pods.Items[pod].Namespace)
                podsOnNode[pods.Items[pod].Name] = pods.Items[pod].Namespace
        }
        return podsOnNode
}

func evictPods(nodeMap map[string]string) {
        fmt.Println(nodeMap)
        for container, namespace := range nodeMap {
                fmt.Println("Container: ", container, "Namespace: ", namespace)
                err := Clientset.CoreV1().Pods(namespace).Delete(container, &metav1.DeleteOptions{})
                if err != nil {
                        fmt.Printf("error removing pod: %v", err)
                }
        }
}

func drainNodes(nodesInAZ []string) {
        /*for i := 0; i < len(nodesInAZ); i ++ {
                fmt.Println(nodesInAZ[i])
                podsOnNode(nodesInAZ[i])
        }*/
        testMap := map[string]string{"gosite-6688c5769b-tdzdj": "gosite"}
        evictPods(testMap)
}

var cordonAZ = &cobra.Command{
        Use: "cordonAZ",
        Short: "Cordon an AZ (only one)",
        Long: "This will make sure the nodes are unschedulable, so when the drain command runs nodes won't go back to the bad az. This also ensurs the cluster autoscaler isn't running",
        Run: func (cmd *cobra.Command, args []string) {
                fmt.Println("Scaling cluster autoscaler to 0 replicas (not really though)")
                //toggleClusterAutoScaler(0)
                fmt.Println("Cordon!\t"+ zone)
                nodes := getNodesInAZ(zone)
                //cordonNodes(nodes)
                drainNodes(nodes)
        },
}
