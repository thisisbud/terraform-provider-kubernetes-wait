package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ tfsdk.DataSourceType = (*kubernetesWaitDataSourceType)(nil)

type kubernetesWaitDataSourceType struct{}

func (d *kubernetesWaitDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `k8-wait todo doc`,

		Attributes: map[string]tfsdk.Attribute{
			"kubernetes_url": {
				Description: "The URL for the request. Supported schemes are `http` and `https`.",
				Type:        types.StringType,
				Required:    true,
			},

			"resource_type": {
				Description: "Todo",
				Type:        types.StringType,
				Optional:    true,
			},

			"resource_name": {
				Description: "Todo",
				Type:        types.StringType,
				Optional:    true,
			},

			"namespace": {
				Description: "Todo",
				Type:        types.StringType,
				Optional:    true,
			},

			"response_body": {
				Description: "The response body returned as a string.",
				Type:        types.StringType,
				Computed:    true,
			},

			"response_headers": {
				Description: `A map of response header field names and values.` +
					` Duplicate headers are concatenated according to [RFC2616](https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2).`,
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Computed: true,
			},

			"status_code": {
				Description: `The HTTP response status code.`,
				Type:        types.Int64Type,
				Computed:    true,
			},

			"id": {
				Description: "The ID of this resource.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

func (d *kubernetesWaitDataSourceType) NewDataSource(context.Context, tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return &kubernetesWaitDataSource{}, nil
}

var _ tfsdk.DataSource = (*kubernetesWaitDataSource)(nil)

type kubernetesWaitDataSource struct{}

func (d *kubernetesWaitDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var model modelV0
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := clientcmd.BuildConfigFromFlags(model.KubernetesURL.Value, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting config from BuildConfigFromFlags",
			fmt.Sprintf("%s", err),
		)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting Kubernetes clientset",
			fmt.Sprintf("%s", err),
		)
		return
	}

	pods, err := clientset.CoreV1().Pods(model.Namespace.Value).List(context.Background(), v1.ListOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting pods",
			fmt.Sprintf("%s", err),
		)
		return
	}

	for _, pod := range pods.Items {
		tflog.Info(ctx, fmt.Sprintf("Pod name: %s\n", pod.Name))
	}

	services, err := clientset.CoreV1().Services(model.Namespace.Value).List(context.Background(), v1.ListOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting services",
			fmt.Sprintf("%s", err),
		)
	}

	for _, s := range services.Items {
		tflog.Info(ctx, fmt.Sprintf("Service name: %s\n", s.Name))
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

type modelV0 struct {
	ID              types.String `tfsdk:"id"`
	KubernetesURL   types.String `tfsdk:"kubernetes_url"`
	ResourceType    types.String `tfsdk:"resource_type"`
	ResourceName    types.String `tfsdk:"resource_name"`
	Namespace       types.String `tfsdk:"namespace"`
	ResponseHeaders types.Map    `tfsdk:"response_headers"`
	ResponseBody    types.String `tfsdk:"response_body"`
	StatusCode      types.Int64  `tfsdk:"status_code"`
}
