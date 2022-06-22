package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var fakeClientset = fake.NewSimpleClientset(
	&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pod-one",
			Namespace:   "namespaceOne",
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{
			NodeName: "nodeNameOne",
			Containers: []v1.Container{
				{
					Name: "pod-one-c-one",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("200m"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("100m"),
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			QOSClass: v1.PodQOSBurstable,
		},
	},
	&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pod-two",
			Namespace:   "myOtherNamespace",
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{
			NodeName: "nodeNameOne",
			Containers: []v1.Container{
				{
					Name: "pod-two-c-one",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("2"),
						},
					},
				},
				{
					Name: "pod-two-c-two",
				},
			},
		},
		Status: v1.PodStatus{
			QOSClass: v1.PodQOSGuaranteed,
		},
	},
)

func TestDummy(t *testing.T) {

	expectedOutput := `
	[
        {
                "namespace":"myOtherNamespace",
                "name":"pod-two",
                "nodeName":"nodeNameOne",
                "qosClass": "Guaranteed",
                "containers": [
                        {
                                "name":"pod-two-c-one",
                                "resources": {
                                        "limits": {
                                                "cpu": "2"
                                        }
                                }
                        },
                        {
                                "name":"pod-two-c-two"
                        }
                ]
        },
        {
                "namespace":"namespaceOne",
                "name":"pod-one",
                "nodeName":"nodeNameOne",
                "qosClass": "Burstable",
                "containers": [
                        {
                                "name":"pod-one-c-one",
                                "resources": {
                                        "limits": {
                                                "cpu": "200m"
                                        },
                                        "requests": {
                                                "cpu": "100m"
                                        }
                                }
                        }
                ]
        }
]`

	template, err := createOutputTemplate("test-template", defaultTemplate)
	if err != nil {
		t.Errorf("Unable to build template %v", err)
	}

	buffer := new(bytes.Buffer)
	ret := showPodInfo("", fakeClientset, template, buffer)
	if ret != nil {
		t.Errorf("showPodInfo failed with: %v", ret)
	}

	ok, err := AreEqualJSON(buffer.String(), expectedOutput)
	if err != nil {
		t.Errorf("Error while trying to check json output: %v", err)
	}

	if !ok {
		t.Errorf("showPodInfo unexpected output:\n\tactual:%v\n\texpected:%v\n", buffer, expectedOutput)
	}

}

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}
