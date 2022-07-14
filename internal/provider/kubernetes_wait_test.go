package provider

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFetchingKubernetesService(t *testing.T) {
	testCases := []struct {
		name               string
		resource           []runtime.Object
		targetNamespace    string
		targetPod          string
		expectedLabelValue string
		expectSuccess      bool
	}{
		{
			name: "existing_service_found",
			resource: []runtime.Object{
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
			resource:        []runtime.Object{},
			targetNamespace: "namespace1",
			targetPod:       "service1",
			expectSuccess:   false,
		},
		{
			name: "wrong namespace",
			resource: []runtime.Object{
				&corev1.Service{
					ObjectMeta: v1.ObjectMeta{
						Name:      "service1",
						Namespace: "namespace2",
					},
				},
			},
			targetNamespace: "namespace1",
			targetPod:       "service1",
			expectSuccess:   false,
		},
		{
			name: "wrong resource type",
			resource: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{
						Name:      "service1",
						Namespace: "namespace1",
					},
				},
			},
			targetNamespace: "namespace1",
			targetPod:       "service1",
			expectSuccess:   false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset(test.resource...)
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
		})
	}
}
