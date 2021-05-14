package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"strings"
)

type Pod struct {
	name, nameSpace string
	writer          io.Writer
}

type Result struct {
	Message string
	Error   error
}

type Node struct {
	name  string
	patch []byte
}

func int64Ptr(i int64) *int64 {
	return &i
}

func toggleClusterAutoScaler(clientset kubernetes.Interface, desiredReplicas int) (*int32, error) {

	currentScale, err := clientset.AppsV1().
		Deployments("kube-system").
		GetScale("cluster-autoscaler", metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	updatedScale := *currentScale
	updatedScale.Spec.Replicas = int32(desiredReplicas)

	_, err = clientset.AppsV1().
		Deployments("kube-system").
		UpdateScale("cluster-autoscaler", &updatedScale)

	if err != nil {
		return nil, err
	}
	return &updatedScale.Spec.Replicas, nil
}

func getNodesInAZ(clientset kubernetes.Interface, zone string) ([]string, error) {
	k8sNodes := make([]string, 0)
	nodes, err := clientset.CoreV1().
		Nodes().
		List(metav1.ListOptions{LabelSelector: "failure-domain.beta.kubernetes.io/zone=" + zone})
	if err != nil {
		return nil, err
	}
	for node, _ := range nodes.Items {
		k8sNodes = append(k8sNodes, nodes.Items[node].ObjectMeta.Labels["kubernetes.io/hostname"])
	}
	if len(k8sNodes) == 0 {
		return nil, err
	}
	return k8sNodes, nil
}

func PatchNode(clientset kubernetes.Interface, node Node, channel chan Result) {
	message := Result{}
	updatedNode, err := clientset.CoreV1().
		Nodes().
		Patch(node.name, "application/strategic-merge-patch+json", node.patch)
	if err != nil {
		message.Error = err
	}

	message.Message = fmt.Sprintf("node: %s unschedulable: %t\n", node.name, updatedNode.Spec.Unschedulable)

	channel <- message
}

func cordonNodes(clientset kubernetes.Interface, nodesInAZ []string, writer io.Writer) bool {

	messages := make(chan Result)
	specPatch := []byte(`{"spec":{"unschedulable":true}}`)
	cordoningErrors := []string{}
	allNodesCordoned := false

	for i := 0; i < len(nodesInAZ); i++ {
		node := Node{
			name:  nodesInAZ[i],
			patch: specPatch,
		}
		go PatchNode(clientset, node, messages)
	}

	for i := 0; i < len(messages); i++ {
		returnMessages := <-messages
		if returnMessages.Error != nil {
			cordoningErrors = append(cordoningErrors, fmt.Sprintf("%v", returnMessages.Error))
		} else {
			fmt.Fprintf(writer, returnMessages.Message)
		}
	}

	if len(cordoningErrors) != 0 {
		fmt.Fprintf(writer, strings.Join(cordoningErrors, "\n"))
	} else {
		allNodesCordoned = true
	}

	return allNodesCordoned

}

func podsOnNode(clientset kubernetes.Interface, nodeName string) (map[string]string, error) {
	podsOnNode := make(map[string]string, 0)
	pods, err := clientset.CoreV1().
		Pods("").
		List(metav1.ListOptions{FieldSelector: "spec.nodeName=" + nodeName})
	if err != nil {
		fmt.Printf("error listing pods on node: %v %v", nodeName, err)
		return nil, err
	}
	for pod, _ := range pods.Items {
		podsOnNode[pods.Items[pod].Name] = pods.Items[pod].Namespace
	}
	return podsOnNode, nil
}

func evictPod(clientset kubernetes.Interface, pod Pod, channel chan Result) {
	message := Result{}
	err := clientset.
		CoreV1().
		Pods(pod.nameSpace).
		Delete(pod.name, &metav1.DeleteOptions{GracePeriodSeconds: int64Ptr(0)})
	if err != nil {
		message.Error = err
	} else {
		message.Message = fmt.Sprintf("evicting pod: %s in the %s namespace", pod.name, pod.nameSpace)
	}
	channel <- message
}

func drainNodes(clientset kubernetes.Interface, nodesInAZ []string) {

	messages := make(chan Result)

	for i := 0; i < len(nodesInAZ); i++ {
		fmt.Println(nodesInAZ[i])
		pods, err := podsOnNode(clientset, nodesInAZ[i])
		if err != nil {
			log.Fatalf("error getting pods on node: %v", err)
		}
		for podName, namespace := range pods {
			p := Pod{
				name:      podName,
				nameSpace: namespace,
				writer:    os.Stdout,
			}
			go evictPod(clientset, p, messages)
		}
		for i := 0; i < len(messages); i++ {
			result := <-messages
			fmt.Println(result.Message)
		}
	}

}

var cordonAZ = &cobra.Command{
	Use:   "cordonAZ",
	Short: "Cordon an AZ (only one)",
	Long:  "This will make sure the nodes are unschedulable, so when the drain command runs nodes won't go back to the bad az. This also ensurs the cluster autoscaler isn't running",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Scaling cluster autoscaler to 0 replicas")
		_, err := toggleClusterAutoScaler(client, 0)
		if err != nil {
			fmt.Printf("error scaling cluster autoscaler: %v", err)
		}
		fmt.Println("Cordoning:\t" + zone)
		nodes, err := getNodesInAZ(client, zone)
		if err != nil {
			log.Fatalf("error getting nodes in %s: %v", zone, err)
		}
		desiredNodesCordoned := cordonNodes(client, nodes, os.Stdout)
		if desiredNodesCordoned {
			drainNodes(client, nodes)
		}
	},
}
