package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLabelUpperCase(t *testing.T) {
	testCases := []struct {
		name               string
		resource           []runtime.Object
		targetNamespace    string
		targetPod          string
		targetLabelKey     string
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

/*
func TestDataSource_NonRegisteredDomainBackoff(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.16",
				Source:            "MehdiAtBud/kubernetes-wait",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `

							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/kubernetes-wait"
									version ="0.1.6"
					  				}
								}
				  			}
							data "kubernetes-wait" "http_test" {
								url = "https://non-existing.thisisbud.com"
								max_elapsed_time = 10
								initial_interval = 100
								multiplier = 1.2
								max_interval = 5000
							}`,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: regexp.MustCompile("no such host"),
			},
		},
	})
}
*/

type TestHttpMock struct {
	server *httptest.Server
}

func setUpMockHttpServer() *TestHttpMock {
	Server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Add("X-Single", "foobar")
			w.Header().Add("X-Double", "1")
			w.Header().Add("X-Double", "2")

			switch r.URL.Path {
			case "/200":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("1.0.0"))
			case "/restricted":
				if r.Header.Get("Authorization") == "Zm9vOmJhcg==" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("1.0.0"))
				} else {
					w.WriteHeader(http.StatusForbidden)
				}
			case "/utf-8/200":
				w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("1.0.0"))
			case "/utf-16/200":
				w.Header().Set("Content-Type", "application/json; charset=UTF-16")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("1.0.0"))
			case "/x509-ca-cert/200":
				w.Header().Set("Content-Type", "application/x-x509-ca-cert")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("pem"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)

	return &TestHttpMock{
		server: Server,
	}
}

//nolint:unparam
func protoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"http": providerserver.NewProtocol6WithError(New()),
	}
}
