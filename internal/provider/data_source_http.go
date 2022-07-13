package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
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
			"initial_interval": {
				Description: "The initial exponential backoff interval.",
				Type:        types.Int64Type,
				Optional:    true,
			},

			"max_elapsed_time": {
				Description: "The maximum time to wait for.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"randomization_factor": {
				Description: "Randomization factor for exponential backoff.",
				Type:        types.StringType,
				Optional:    true,
			},
			"multiplier": {
				Description: "Multiplier for exponential backoff.",
				Type:        types.StringType,
				Optional:    true,
			},
			"max_interval": {
				Description: "Maximum interval factor for exponential backoff.",
				Type:        types.Int64Type,
				Optional:    true,
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

	var randomization_factor, multiplier float64

	randomization_factor = backoff.DefaultRandomizationFactor
	multiplier = backoff.DefaultMultiplier

	if model.InitialInterval.Value == 0 {
		model.InitialInterval.Value = int64(backoff.DefaultInitialInterval)
	}

	if model.MaxElapsedTime.Value == 0 {
		model.MaxElapsedTime.Value = int64(backoff.DefaultMaxElapsedTime)
	}

	if model.MaxElapsedTime.Value == 0 {
		model.MaxInterval.Value = int64(backoff.DefaultMaxElapsedTime)
	}

	if len(model.RandomizationFactor.Value) > 0 {
		randomization_factor, err = strconv.ParseFloat(model.RandomizationFactor.Value, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"error converting string to float64",
				fmt.Sprintf("%s", err),
			)
			return
		}
	}

	if len(model.Multiplier.Value) > 0 {
		multiplier, err = strconv.ParseFloat(model.Multiplier.Value, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"error converting string to float64",
				fmt.Sprintf("%s", err),
			)
			return
		}
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(model.MaxElapsedTime.Value) * time.Second
	b.InitialInterval = time.Duration(model.InitialInterval.Value) * time.Millisecond
	b.RandomizationFactor = randomization_factor
	b.Multiplier = multiplier
	b.MaxInterval = time.Duration(model.MaxInterval.Value) * time.Millisecond

	s, err := json.MarshalIndent(b, "", "   ")
	tflog.Info(ctx, fmt.Sprintf("Backoff configuration :  %s", s))

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Json.MarshalIndent error :  %s", err))
	}

	retries := 0
	err = backoff.Retry(func() error {
		tflog.Info(ctx, "Retrieving services from Kubernets cluster")
		requestResource(ctx, clientset, model.Namespace.Value, model.ResourceName.Value)
		_, err := clientset.CoreV1().Services(model.Namespace.Value).Get(ctx, model.ResourceName.Value, v1.GetOptions{})
		tflog.Info(ctx, fmt.Sprintf("Number of retries %d", retries))
		retries++
		return err
	}, b)
	if err != nil {
		resp.Diagnostics.AddError(
			"error getting services",
			fmt.Sprintf("%s", err),
		)
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

type modelV0 struct {
	ID                  types.String `tfsdk:"id"`
	KubernetesURL       types.String `tfsdk:"kubernetes_url"`
	ResourceType        types.String `tfsdk:"resource_type"`
	ResourceName        types.String `tfsdk:"resource_name"`
	Namespace           types.String `tfsdk:"namespace"`
	ResponseHeaders     types.Map    `tfsdk:"response_headers"`
	ResponseBody        types.String `tfsdk:"response_body"`
	StatusCode          types.Int64  `tfsdk:"status_code"`
	InitialInterval     types.Int64  `tfsdk:"initial_interval"`
	MaxElapsedTime      types.Int64  `tfsdk:"max_elapsed_time"`
	RandomizationFactor types.String `tfsdk:"randomization_factor"`
	Multiplier          types.String `tfsdk:"multiplier"`
	MaxInterval         types.Int64  `tfsdk:"max_interval"`
}

func requestResource(ctx context.Context, clientset kubernetes.Interface, namespace, resourceName string) error {
	_, err := clientset.CoreV1().Services(namespace).Get(ctx, resourceName, v1.GetOptions{})
	return err
}
