package provider

import (
	"context"
	"fmt"
	"os"

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
				Description:         "The Kubernetes URL. Can be sourced from KUBERNETES_URL.",
				MarkdownDescription: "The Kubernetes URL. Can be sourced from `KUBERNETES_URL`.",
				Optional:            true,
				Computed:            true,
			},
		},
	}, nil
}

type providerData struct {
	Host types.String `tfsdk:"host"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var config providerData
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

	if config.Host.Null {
		host = os.Getenv("KUBERNETES_URL")
	} else {
		host = config.Host.Value
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Unable to find host",
			"KUBERNETES_URL cannot be an empty string",
		)
		return
	}

	configk8, err := clientcmd.BuildConfigFromFlags(host, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting config from BuildConfigFromFlags",
			fmt.Sprintf("%s", err),
		)
		return
	}

	clientset, err := kubernetes.NewForConfig(configk8)
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
