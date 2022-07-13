package provider

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLabelUpperCase(t *testing.T) {
	testCases := []struct {
		name               string
		pods               []runtime.Object
		targetNamespace    string
		targetPod          string
		targetLabelKey     string
		expectedLabelValue string
		expectSuccess      bool
	}{
		{
			name: "existing_service_found",
			pods: []runtime.Object{
				&corev1.Service{
					ObjectMeta: v1.ObjectMeta{
						Name:      "service1",
						Namespace: "namespace1",
					},
				},
			},
			targetNamespace: "namespace1",
			targetPod:       "service1",
			expectSuccess:   true,
		},
		{
			name:            "no_service_existing",
			pods:            []runtime.Object{},
			targetNamespace: "namespace1",
			targetPod:       "service1",
			expectSuccess:   false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset(test.pods...)
			err := requestResource(
				context.Background(),
				fakeClientset,
				test.targetNamespace,
				test.targetPod,
			)
			if err != nil && test.expectSuccess {
				t.Fatalf("unexpected error getting label: %v", err)
			} else if err == nil && !test.expectSuccess {
				t.Fatalf("expected error but received none getting label")
			}
			// } else if labelValue != test.expectedLabelValue && test.expectSuccess {
			// 	t.Fatalf("label value %s unexpectedly not equal to %s", labelValue, test.expectedLabelValue)
			// } else if labelValue == test.expectedLabelValue && !test.expectSuccess {
			// 	t.Fatalf("label values are unexpectedly equal: %s", labelValue)
			// }
		})
	}
}
