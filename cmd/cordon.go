package cmd

import (
        "fmt"
        "github.com/spf13/cobra"
        "k8s.io/client-go/kubernetes"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "log"
        "io"
        "os"
        //scalev1 "k8s.io/api/autoscaling/v1"
//        "errors"
)

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

//func getNodesInAZ(zone string) ([]string, error) {
func getNodesInAZ(clientset kubernetes.Interface, zone string) ([]string, error) {
        k8sNodes := make([]string, 0)
        nodes, err := clientset.CoreV1().
            Nodes().
            List(metav1.ListOptions{LabelSelector: "failure-domain.beta.kubernetes.io/zone="+zone})
        if err != nil {
                //fmt.Printf("error listing nodes in az: %v", err)
                return nil, err
        }
        for node, _ := range nodes.Items {
                k8sNodes = append(k8sNodes, nodes.Items[node].ObjectMeta.Labels["kubernetes.io/hostname"])
        }
        if len(k8sNodes) == 0 {
                //return k8sNodes, errors.New("There are no nodes in the given zone")
                k8sNodes = append(k8sNodes, "docker-desktop")
                return nil, err
        }
        return k8sNodes, err
}

func cordonNodes(clientset kubernetes.Interface, nodesInAZ []string, writer io.Writer) {
        patch := []byte(`{"spec":{"unschedulable":true}}`)
        for i := 0; i < len(nodesInAZ); i ++ {
                _, err := clientset.CoreV1().
                    Nodes().
                    Patch(nodesInAZ[i], "application/strategic-merge-patch+json", patch)
                if err != nil {
                        log.Fatalf("error patching node: %v", err)
                }
                message := fmt.Sprintf("Successfully Cordoned node: %s \n", nodesInAZ[i])
                fmt.Fprintf(writer, message)
        }
}

func podsOnNode(clientset kubernetes.Interface, nodeName string) (map[string]string, error) {
        podsOnNode := make(map[string]string, 0)
        pods, err := clientset.CoreV1().
            Pods("").
            List(metav1.ListOptions{FieldSelector: "spec.nodeName="+nodeName})
        if err != nil {
                fmt.Printf("error listing pods on node: %v %v", nodeName, err)
                return nil, err
        }
        for pod, _ := range pods.Items {
                //fmt.Println(pods.Items[pod].Name, pods.Items[pod].Namespace)
                podsOnNode[pods.Items[pod].Name] = pods.Items[pod].Namespace
        }
        return podsOnNode, nil
}

func evictPods(clientset kubernetes.Interface, podMap map[string]string, writer io.Writer) {
        for container, namespace := range podMap {
                fmt.Println("Container: ", container, "Namespace: ", namespace)
                err := clientset.
                        CoreV1().
                        Pods(namespace).
                        Delete(container, &metav1.DeleteOptions{})
                if err != nil {
                        log.Fatalf("error removing pod: %v", err)
                }
                messge := fmt.Sprintf("evicting pod: %s in the %s namespace", container, namespace)
                fmt.Fprintf(writer, messge)
        }
}

func drainNodes(clientset kubernetes.Interface, nodesInAZ []string) {
        affectedPods := make(map[string]string, 0)

        for i := 0; i < len(nodesInAZ); i ++ {
                fmt.Println(nodesInAZ[i])
                pods, err := podsOnNode(clientset, nodesInAZ[i])
                if err != nil {
                        log.Fatalf("error getting pods on node: %v", err)
                }
                for podName, namespace := range pods {
                        affectedPods[podName] = namespace
                }
        }

        evictPods(clientset, affectedPods, os.Stdout)
}

var cordonAZ = &cobra.Command{
        Use: "cordonAZ",
        Short: "Cordon an AZ (only one)",
        Long: "This will make sure the nodes are unschedulable, so when the drain command runs nodes won't go back to the bad az. This also ensurs the cluster autoscaler isn't running",
        Run: func (cmd *cobra.Command, args []string) {
                fmt.Println("Scaling cluster autoscaler to 0 replicas")
                _, err := toggleClusterAutoScaler(client, 0)
                if err != nil {
                        fmt.Printf("error scaling cluster autoscaler: %v", err)
                }
                fmt.Println("Cordon!\t"+ zone)
                nodes, err := getNodesInAZ(client, zone)
                if err != nil {
                        log.Fatalf("error getting nodes in %s: %v", zone, err)
                }
                cordonNodes(client, nodes, os.Stdout)
                drainNodes(client, nodes)
        },
}
