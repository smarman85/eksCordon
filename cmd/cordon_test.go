package cmd

import (
        "testing"
        "k8s.io/client-go/kubernetes/fake"
        "k8s.io/client-go/kubernetes"
        v1 "k8s.io/api/core/v1"
        appv1 "k8s.io/api/apps/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TesttoggleClusterAutoScaler(t *testing.T) {

        currentReplicas := *int32(1)
        desiredReplicas := *int32(0)

        clientset := fake.NewSimpleClientset(&appv1.Deployment{
                ObjectMeta: metav1.ObjectMeta{
                        Name: "cluster-autoscaler",
                        Namespace: "kube-system",
                },
                Spec: appv1.DeploymentSpec{
                        Replicas: currentReplicas,
                },
        })

        got, err := toggleClusterAutoScaler(clientset, 0)
        if err != nil {
                t.Fatalf("errored toggeling cas: %q", err)
        }
        want := fake.NewSimpleClientset(&appv1.Deployment{
                ObjectMeta: metav1.ObjectMeta{
                        Name: "cluster-autoscaler",
                        Namespace: "kube-system",
                },
                Spec: appv1.DeploymentSpec{
                        Replicas: desiredReplicas,
                },
        })

        if got != want {
                t.Errorf("got %q want %q", got, want)
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
                               got := getNodesInAZ(single.clientset, single.zone)
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
