package cmd

import (
	"bytes"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
)

func int32Ptr(i int32) *int32 {
	return &i
}

func TesttoggleClusterAutoScaler(t *testing.T) {

	clientset := fake.NewSimpleClientset(&appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-autoscaler",
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
								"kubernetes.io/hostname":                 "goodhost",
							},
						},
					}, v1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: "host2",
							Labels: map[string]string{
								"failure-domain.beta.kubernetes.io/zone": "us-east-1c",
								"kubernetes.io/hostname":                 "badhost",
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
			want      []string
		}) func(t *testing.T) {
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

func TestPatchNode(t *testing.T) {
	clientset := fake.NewSimpleClientset(&v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "host1",
			Labels: map[string]string{
				"failure-domain.beta.kubernetes.io/zone": "us-east-1a",
				"kubernetes.io/hostname":                 "host1",
			},
		},
	})

	messages := make(chan Result)
	node := Node{
		name:  "host1",
		patch: []byte(`{"spec":{"unschedulable":true}}`),
	}

	go PatchNode(clientset, node, messages)

	var got Result
	select {
	case got = <-messages:
		fmt.Println(got)
	}
	t.Log("t.Log:", got)
	if got.Error != nil {
		t.Errorf("got %q", got.Error)
	}
}

func NewControllerRef(owner metav1.ObjectMeta, gvk schema.GroupVersionKind) *metav1.OwnerReference {
	blockOwnerDeletion := true
	isController := true
	return &metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}

func TestPodsOnNode(t *testing.T) {

	rsgvk := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "ReplicaSet",
	}

	dsgvk := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "DaemonSet",
	}

	rs := metav1.ObjectMeta{
		UID:  "uid1",
		Name: "ms-rs",
	}

	ds := metav1.ObjectMeta{
		UID:  "uid2",
		Name: "m-ds",
	}

	clientset := fake.NewSimpleClientset(
		&v1.PodList{
			Items: []v1.Pod{
				v1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind: "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mr-meeseeks",
						Namespace: "existence",
						OwnerReferences: []metav1.OwnerReference{
							*NewControllerRef(rs, rsgvk),
						},
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
				},
				v1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind: "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "morty",
						Namespace: "existence",
						OwnerReferences: []metav1.OwnerReference{
							*NewControllerRef(ds, dsgvk),
						},
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
				},
				v1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind: "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "jerry",
						Namespace: "existence",
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
				},
				v1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind: "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rick",
						Namespace: "existence",
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
						Volumes: []v1.Volume{
							v1.Volume{
								Name: "ram-drvie",
								VolumeSource: v1.VolumeSource{
									EmptyDir: &v1.EmptyDirVolumeSource{
										Medium: "Memory",
									},
								},
							},
						},
					},
				},
				v1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind: "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "snowball",
						Namespace: "existence",
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
						Volumes: []v1.Volume{
							v1.Volume{
								Name: "bifrost",
								VolumeSource: v1.VolumeSource{
									EmptyDir: &v1.EmptyDirVolumeSource{
										Medium: "",
									},
								},
							},
						},
					},
				},
			},
		})

	nodeName := "node1"

	got, err := podsOnNode(clientset, nodeName)
	if err != nil {
		t.Errorf("Error getting pod list: %v", err)
	}
	want := map[string]string{
		"mr-meeseeks": "existence",
		"jerry":       "existence",
		"snowball":    "existence",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestEvictPod(t *testing.T) {
	buffer := bytes.Buffer{}

	messages := make(chan Result)
	clientset := fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-space",
			},
		},
	)
	pod := Pod{
		name:      "test-pod",
		nameSpace: "test-space",
		writer:    &buffer,
	}

	go evictPod(clientset, pod, messages)
	var got Result
	select {
	case got = <-messages:
		fmt.Println(got)
	}

	t.Log("t.Log:", got)
	if got.Error != nil {
		t.Errorf("got %q", got.Error)
	}
}
