// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/matrix-org/gomatrix"
)

// Ensure MatrixProvider satisfies various provider interfaces.
var _ provider.Provider = &MatrixProvider{}

// MatrixProvider defines the provider implementation.
type MatrixProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// MatrixProviderModel describes the provider data model.
type MatrixProviderModel struct {
	ClientServerUrl    types.String `tfsdk:"client_server_url"`
	DefaultAccessToken types.String `tfsdk:"default_access_token"`
	DefaultUserID      types.String `tfsdk:"default_user_id"`
}

func (p *MatrixProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "matrix"
	resp.Version = p.version
}

func (p *MatrixProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_server_url": schema.StringAttribute{
				MarkdownDescription: "Address of the matrix server you are acting upon",
				Required:            true,
			},
			"default_access_token": schema.StringAttribute{
				MarkdownDescription: "The default access token to use for things like content uploads.",
				Required:            true,
				Sensitive:           true,
			},
			"default_user_id": schema.StringAttribute{
				MarkdownDescription: "The default user id to use for things like content uploads. This must match the access_token",
				Required:            true,
			},
		},
	}
}

func (p *MatrixProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config MatrixProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.ClientServerUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_server_url"),
			"Unknown Matrix Server URL",
			"The provider cannot create the Matrix API client as there is an unknown configuration value for the Matrix API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MATRIX_CLIENT_SERVER_URL environment variable.",
		)
	}

	if config.DefaultAccessToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_access_token"),
			"Unknown Default Access Token",
			"The provider cannot create the Matrix API client as there is an unknown configuration value for the default AccessToken. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MATRIX_DEFAULT_ACCESS_TOKEN environment variable.",
		)
	}

	if config.DefaultUserID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_user_id"),
			"Unknown Default User ID",
			"The provider cannot create the Matrix API client as there is an unknown configuration value for the default UserID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MATRIX_DEFAULT_USERID environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	client_server_url := os.Getenv("MATRIX_CLIENT_SERVER_URL")
	default_access_token := os.Getenv("MATRIX_DEFAULT_ACCESS_TOKEN")
	default_user_id := os.Getenv("MATRIX_DEFAULT_USERID")

	if !config.ClientServerUrl.IsNull() {
		client_server_url = config.ClientServerUrl.ValueString()
	}

	if !config.DefaultAccessToken.IsNull() {
		default_access_token = config.DefaultAccessToken.ValueString()
	}

	if !config.DefaultUserID.IsNull() {
		default_user_id = config.DefaultUserID.ValueString()
	}

	if client_server_url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_server_url"),
			"Missing Matrix Server URL",
			"The provider cannot create the Matrix API client as there is a missing or empty value for the Matrix API host. "+
				"Set the client_server_url value in the configuration or use the MATRIX_CLIENT_SERVER_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if default_access_token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_access_token"),
			"Missing Default Access Token",
			"The provider cannot create the Matrix API client as there is a missing or empty value for the default AccessToken. "+
				"Set the default_access_token value in the configuration or use the MATRIX_DEFAULT_ACCESS_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if default_user_id == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("default_user_id"),
			"Missing Default UserID",
			"The provider cannot create the Matrix API client as there is a missing or empty value for the default UserID. "+
				"Set the default_user_id value in the configuration or use the MATRIX_DEFAULT_USERID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "client_server_url", client_server_url)
	ctx = tflog.SetField(ctx, "default_access_token", default_access_token)
	ctx = tflog.SetField(ctx, "default_user_id", default_user_id)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "default_access_token")

	tflog.Debug(ctx, "Creating Matrix client")

	// Example client configuration for data sources and resources
	client, err := gomatrix.NewClient(client_server_url, default_user_id, default_access_token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Matrix API Client",
			"An unexpected error occurred when creating the Matrix API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Matrix Client Error: "+err.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Matrix client", map[string]any{"success": true})
}

func (p *MatrixProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *MatrixProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MatrixProvider{
			version: version,
		}
	}
}
