package cmd

import (
        "fmt"
        "testing"
        "k8s.io/client-go/kubernetes/fake"
        "k8s.io/client-go/kubernetes"
        "reflect"
        "bytes"
        v1 "k8s.io/api/core/v1"
        appv1 "k8s.io/api/apps/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        //scalev1 "k8s.io/api/autoscaling/v1"
)

func int32Ptr(i int32) *int32 {
        return &i
}

func TesttoggleClusterAutoScaler(t *testing.T) {

        clientset := fake.NewSimpleClientset(&appv1.Deployment{
                ObjectMeta: metav1.ObjectMeta{
                        Name: "cluster-autoscaler",
                        Namespace: "kube-system",
                },
                Spec: appv1.DeploymentSpec{
                        Replicas: int32Ptr(1),
                },
        })

        got, err := toggleClusterAutoScaler(clientset, 0)
        fmt.Println(got)
        if err != nil {
                t.Fatalf("errored toggeling cas: %q", err)
        }
        //want := fake.NewSimpleClientset(&appv1.Deployment{
        want := int32Ptr(0)

        if got != want {
                t.Errorf("got %p want %p", got, want)
        }

}

func TestGetNodesInAZ(t *testing.T) {
       data := []struct {
               clientset kubernetes.Interface
               zone      string
               want      []string
       }{
               {
                       clientset: fake.NewSimpleClientset(&v1.NodeList{
                               Items: []v1.Node{
                                       v1.Node{
                                               ObjectMeta: metav1.ObjectMeta{
                                                       Name: "host1",
                                                       Labels: map[string]string{
                                                               "failure-domain.beta.kubernetes.io/zone": "us-east-1a",
                                                               "kubernetes.io/hostname": "goodhost",
                                                       },
                                               },
                                       }, v1.Node{
                                               ObjectMeta: metav1.ObjectMeta{
                                                       Name: "host2",
                                                       Labels: map[string]string{
                                                               "failure-domain.beta.kubernetes.io/zone": "us-east-1c",
                                                               "kubernetes.io/hostname": "badhost",
                                                       },
                                               },
                                       },
                               },
                       }),
                       zone: "us-east-1a",
                       want: []string{"goodhost"},
               },
       }

       for _, single := range data {
               t.Run("", func(single struct {
                       clientset kubernetes.Interface
                       zone      string
                       want     []string
               }) func (t *testing.T) {
                       return func(t *testing.T) {
                               got, err := getNodesInAZ(single.clientset, single.zone)
                               if err != nil {
                                       t.Errorf("error getting nodes in %s: %v", single.zone, err)
                               }
                               if len(got) != len(single.want) {
                                       t.Errorf("slices are unequal lengths got %q want %q", got, single.want)
                               }
                               for i, v := range got {
                                       if v != single.want[i] {
                                              t.Errorf("wtf got %q want %q", got, single.want)
                                       }
                               }
                               for _, v := range got {
                                       if v == "badhost" {
                                              t.Errorf("badhost should not be returned got %q want %q", got, single.want)
                                       }
                               }
                       }
               }(single))
       }
}

func TestCordonNodes(t *testing.T) {
        clientset := fake.NewSimpleClientset(&v1.NodeList{
                Items: []v1.Node{
                        v1.Node{
                                ObjectMeta: metav1.ObjectMeta{
                                        Name: "host1",
                                        Labels: map[string]string{
                                                "failure-domain.beta.kubernetes.io/zone": "us-east-1a",
                                                "kubernetes.io/hostname": "host1",
                                        },
                                },
                        }, v1.Node{
                                ObjectMeta: metav1.ObjectMeta{
                                        Name: "host2",
                                        Labels: map[string]string{
                                                "failure-domain.beta.kubernetes.io/zone": "us-east-1a",
                                                "kubernetes.io/hostname": "host2",
                                        },
                                },
                        },
                },
        })

        testHosts := []string{"host1", "host2"}
        buffer := bytes.Buffer{}

        cordonNodes(clientset, testHosts, &buffer)
        got := buffer.String()
        want := "Successfully Cordoned node: host1 \nSuccessfully Cordoned node: host2 \n"
        if got != want {
                t.Errorf("got %q want %q", got, want)
        }
}

func TestPodsOnNode(t *testing.T) {
        clientset := fake.NewSimpleClientset(
                &v1.PodList{
                        Items: []v1.Pod{
                                v1.Pod{
                                        ObjectMeta: metav1.ObjectMeta{
                                                Name: "mr-meeseeks",
                                                Namespace: "existence",
                                        },
                                        Spec: v1.PodSpec{
                                                NodeName: "node1",
                                        },
                                },
                        },
                },
        )
        nodeName := "node1"

        got, err :=  podsOnNode(clientset, nodeName)
        if err != nil {
                t.Errorf("Error getting pod list: %v", err)
        }
        want := map[string]string{"mr-meeseeks": "existence"}

        if !reflect.DeepEqual(got, want) {
                t.Errorf("got %q want %q", got, want)
        }
}

func TestEvictPods(t *testing.T) {
        clientset := fake.NewSimpleClientset(
                &v1.Pod{
                        ObjectMeta: metav1.ObjectMeta{
                                Name: "test-pod",
                                Namespace: "test-space",
                        },
                },
        )
        podMap := map[string]string{"test-pod": "test-space"}

        buffer := bytes.Buffer{}
        evictPods(clientset, podMap, &buffer)
        got := buffer.String()
        want := "evicting pod: test-pod in the test-space namespace"

        if got != want {
                t.Errorf("got %q want %q", got, want)
        }
}
