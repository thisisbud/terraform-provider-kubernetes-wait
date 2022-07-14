package provider

import (
	"context"
	"encoding/pem"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ tfsdk.Provider = (*provider)(nil)

func (p *provider) GetResources(context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{}, nil
}

func (p *provider) GetDataSources(context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"kubernetes-wait": &kubernetesWaitDataSourceType{},
	}, nil
}

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     kubernetes.Clientset
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				Type:                types.StringType,
				Description:         "The Kubernetes host URL",
				MarkdownDescription: "The Kubernetes host URL.",
				Required:            true,
			},
			"cluster_ca_certificate": {
				Type:                types.StringType,
				Description:         "PEM-encoded root certificates bundle for TLS authentication.",
				MarkdownDescription: "PEM-encoded root certificates bundle for TLS authentication.",
				Required:            true,
			},
		},
	}, nil
}

type providerData struct {
	Host                 types.String `tfsdk:"host"`
	ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
	Token                types.String `tfsdk:"token"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var config providerData
	overrides := &clientcmd.ConfigOverrides{}
	loader := &clientcmd.ClientConfigLoadingRules{}

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var host string
	if config.Host.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as kubernetes url",
		)
	}

	if !config.Host.IsNull() && !config.Host.IsUnknown() {
		host = config.Host.Value
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Unable to find host",
			"Host cannot be an empty string",
		)
		return
	}

	var clusterCaCertificate string
	if !config.ClusterCACertificate.IsNull() && !config.ClusterCACertificate.IsUnknown() {
		clusterCaCertificate = config.ClusterCACertificate.Value
	}

	ca, _ := pem.Decode([]byte(clusterCaCertificate))
	if ca == nil || ca.Type != "CERTIFICATE" {
		resp.Diagnostics.AddError(
			"Invalid attribute in provider configuration",
			"'cluster_ca_certificate' is not a valid PEM encoded certificate",
		)
	}

	var token string
	if !config.Token.IsNull() && !config.Token.IsUnknown() {
		token = config.Token.Value
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"Token cannot be an empty string",
		)
		return
	}

	overrides.ClusterInfo.Server = host
	overrides.ClusterInfo.CertificateAuthorityData = []byte(clusterCaCertificate)
	overrides.AuthInfo.Token = token

	configk8 := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	cc, err := configk8.ClientConfig()
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting config from ClientConfig",
			fmt.Sprintf("%s", err),
		)
		return
	}

	clientset, err := kubernetes.NewForConfig(cc)
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting Kubernetes clientset",
			fmt.Sprintf("%s", err),
		)
		return
	}

	p.client = *clientset
	p.configured = true
}
