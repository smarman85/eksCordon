package cmd

import (
	"bytes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
)

func TestDisplayAvailabilityZones(t *testing.T) {

	t.Run("Prints AZs and number of nodes in an az", func(t *testing.T) {
		buffer := bytes.Buffer{}
		zones := map[string]int{
			"us-east-1a": 17,
		}
		DisplayAvailabilityZones(&buffer, zones)

		got := buffer.String()
		want := "AZ: us-east-1a\tNumber of nodes: 17\n"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

}

func TestGetFailureDomains(t *testing.T) {
	data := []struct {
		clientset kubernetes.Interface
		err       error
		want      map[string]int
	}{
		{
			clientset: fake.NewSimpleClientset(&v1.NodeList{
				Items: []v1.Node{v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"failure-domain.beta.kubernetes.io/zone": "us-east-1a",
						},
					}},
				},
			}),
			want: map[string]int{"us-east-1a": 1},
		},
	}

	for _, single := range data {
		t.Run("", func(single struct {
			clientset kubernetes.Interface
			err       error
			want      map[string]int
		}) func(t *testing.T) {
			return func(t *testing.T) {

				availabilitZones, err := GetFailureDomains(single.clientset)
				if err != nil {
					t.Fatalf(err.Error())
				}
				if !reflect.DeepEqual(availabilitZones, single.want) {
					t.Errorf("got %q want %q", availabilitZones, single.want)
				}
			}
		}(single))
	}

}
