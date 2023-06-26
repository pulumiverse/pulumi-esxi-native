// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"fmt"
	"github.com/edmondshtogu/pulumi-esxi-native/provider/pkg/provider/esxi"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"google.golang.org/protobuf/types/known/emptypb"
	"math/rand"
	"os"
	"time"

	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	pbempty "github.com/golang/protobuf/ptypes/empty"
)

type cancellationContext struct {
	context context.Context
	cancel  context.CancelFunc
}

func makeCancellationContext() *cancellationContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &cancellationContext{
		context: ctx,
		cancel:  cancel,
	}
}

type cfnProvider struct {
	pulumirpc.UnimplementedResourceProviderServer

	host     *provider.HostClient
	name     string
	canceler *cancellationContext

	configured bool
	version    string

	pulumiSchema []byte

	esxi *esxi.Host
}

var _ pulumirpc.ResourceProviderServer = (*cfnProvider)(nil)

func newESXiNativeProvider(host *provider.HostClient, name, version string, pulumiSchema []byte) (
	pulumirpc.ResourceProviderServer, error) {
	return &cfnProvider{
		host:         host,
		canceler:     makeCancellationContext(),
		name:         name,
		version:      version,
		pulumiSchema: pulumiSchema,
	}, nil
}

// Attach sends the engine address to an already running plugin.
func (p *cfnProvider) Attach(context context.Context, req *pulumirpc.PluginAttach) (*emptypb.Empty, error) {
	host, err := provider.NewHostClient(req.GetAddress())
	if err != nil {
		return nil, err
	}
	p.host = host
	return &pbempty.Empty{}, nil
}

// Call dynamically executes a method in the provider associated with a component resource.
func (p *cfnProvider) Call(ctx context.Context, req *pulumirpc.CallRequest) (*pulumirpc.CallResponse, error) {
	return nil, status.Error(codes.Unimplemented, "call is not yet implemented")
}

// Construct creates a new component resource.
func (p *cfnProvider) Construct(ctx context.Context, req *pulumirpc.ConstructRequest) (*pulumirpc.ConstructResponse, error) {
	return nil, status.Error(codes.Unimplemented, "construct is not yet implemented")
}

// CheckConfig validates the configuration for this provider.
func (p *cfnProvider) CheckConfig(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	// TODO: here we validate the connection
	return &pulumirpc.CheckResponse{Inputs: req.GetNews()}, nil
}

// DiffConfig diffs the configuration for this provider.
func (p *cfnProvider) DiffConfig(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.DiffConfig(%s)", p.name, urn)
	glog.V(9).Infof("%s executing", label)

	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.olds", label),
		KeepUnknowns: true,
	})
	if err != nil {
		return nil, err
	}
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.news", label),
		KeepUnknowns: true,
		RejectAssets: true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "diffConfig failed because of malformed resource inputs")
	}

	diff := olds.Diff(news)
	if diff == nil {
		return &pulumirpc.DiffResponse{Changes: pulumirpc.DiffResponse_DIFF_NONE}, nil
	}

	var diffs, replaces []string
	for _, k := range diff.Keys() {
		diffs = append(diffs, string(k))
	}

	return &pulumirpc.DiffResponse{
		Changes:  pulumirpc.DiffResponse_DIFF_SOME,
		Diffs:    diffs,
		Replaces: replaces,
	}, nil
}

// Configure configures the resource provider with "globals" that control its behavior.
func (p *cfnProvider) Configure(_ context.Context, req *pulumirpc.ConfigureRequest) (*pulumirpc.ConfigureResponse, error) {
	vars := req.GetVariables()

	var host, user, pass, sshPort, sslPort, ovfLoc string

	if v, ok := varsOrEnv(vars, "esxi-native:config:host", "ESXI_HOST"); ok {
		host = v
	}
	if v, ok := varsOrEnv(vars, "esxi-native:config:username", "ESXI_USERNAME"); ok {
		user = v
	}
	if v, ok := varsOrEnv(vars, "esxi-native:config:password", "ESXI_PASSWORD"); ok {
		pass = v
	}
	if v, ok := varsOrEnv(vars, "esxi-native:config:sshPort", "ESXI_SSH_PORT"); ok {
		sshPort = v
	}
	if v, ok := varsOrEnv(vars, "esxi-native:config:sslPort", "ESXI_SSL_PORT"); ok {
		sslPort = v
	}
	if v, ok := varsOrEnv(vars, "esxi-native:config:ovfToolLocation", "ESXI_OVFTOOL_LOCATION"); ok {
		ovfLoc = v
	}

	if len(host) > 0 || len(user) > 0 || len(pass) > 0 || len(sslPort) > 0 || len(sshPort) > 0 || len(ovfLoc) > 0 {
		// If all required values are not present/valid, the client will return an appropriate error.
		esxi := esxi.NewHost(host, sshPort, sslPort, user, pass, ovfLoc)
		p.esxi = &esxi
	}

	err := p.esxi.ValidateCreds()
	if err != nil {
		return nil, err
	}

	p.configured = true

	return &pulumirpc.ConfigureResponse{
		AcceptSecrets: true,
	}, nil
}

// Invoke dynamically executes a built-in function in the provider.
func (p *cfnProvider) Invoke(_ context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	tok := req.GetTok()
	return nil, fmt.Errorf("unknown Invoke token '%s'", tok)
}

// StreamInvoke dynamically executes a built-in function in the provider. The result is streamed
// back as a series of messages.
func (p *cfnProvider) StreamInvoke(req *pulumirpc.InvokeRequest, server pulumirpc.ResourceProvider_StreamInvokeServer) error {
	tok := req.GetTok()
	return fmt.Errorf("unknown StreamInvoke token '%s'", tok)
}

// Check validates that the given property bag is valid for a resource of the given type and returns
// the inputs that should be passed to successive calls to Diff, Create, or Update for this
// resource. As a rule, the provider inputs returned by a call to Check should preserve the original
// representation of the properties as present in the program inputs. Though this rule is not
// required for correctness, violations thereof can negatively impact the end-user experience, as
// the provider inputs are using for detecting and rendering diffs.
func (p *cfnProvider) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Create(%s)", p.name, urn)
	logging.V(9).Infof("%s executing", label)

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

// Diff checks what impacts a hypothetical update will have on the resource's properties.
func (p *cfnProvider) Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Diff(%s)", p.name, urn)
	logging.V(9).Infof("%s executing", label)

	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	d := olds.Diff(news)
	changes := pulumirpc.DiffResponse_DIFF_NONE

	// Replace the below condition with logic specific to your provider
	if d.Changed("length") {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:  changes,
		Replaces: []string{"length"},
	}, nil
}

// Create allocates a new instance of the provided resource and returns its unique ID afterwards.
func (p *cfnProvider) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Create(%s)", p.name, urn)
	logging.V(9).Infof("%s executing", label)

	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	// Replace the below random number implementation with logic specific to your provider
	if !inputs["length"].IsNumber() {
		return nil, fmt.Errorf("expected input property 'length' of type 'number' but got '%s", inputs["length"].TypeString())
	}

	n := int(inputs["length"].NumberValue())

	// Actually "create" the random number
	result := makeRandom(n)

	outputs := map[string]interface{}{
		"length": n,
		"result": result,
	}

	outputProperties, err := plugin.MarshalProperties(
		resource.NewPropertyMapFromMap(outputs),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CreateResponse{
		Id:         result,
		Properties: outputProperties,
	}, nil
}

// Read the current live state associated with a resource.
func (p *cfnProvider) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Read(%s)", p.name, urn)
	logging.V(9).Infof("%s executing", label)
	msg := fmt.Sprintf("Read is not yet implemented for %s", urn.Type())
	return nil, status.Error(codes.Unimplemented, msg)
}

// Update updates an existing resource with new values.
func (p *cfnProvider) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Update(%s)", p.name, urn)
	logging.V(9).Infof("%s executing", label)
	// Our example Random resource will never be updated - if there is a diff, it will be a replacement.
	msg := fmt.Sprintf("Update is not yet implemented for %s", urn.Type())
	return nil, status.Error(codes.Unimplemented, msg)
}

// Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed
// to still exist.
func (p *cfnProvider) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Update(%s)", p.name, urn)
	logging.V(9).Infof("%s executing", label)
	// Implement Delete logic specific to your provider.
	// Note that for our Random resource, we don't have to do anything on Delete.
	return &pbempty.Empty{}, nil
}

// GetPluginInfo returns generic information about this plugin, like its version.
func (p *cfnProvider) GetPluginInfo(context.Context, *pbempty.Empty) (*pulumirpc.PluginInfo, error) {
	return &pulumirpc.PluginInfo{
		Version: p.version,
	}, nil
}

// GetSchema returns the JSON-serialized schema for the provider.
func (p *cfnProvider) GetSchema(ctx context.Context, req *pulumirpc.GetSchemaRequest) (*pulumirpc.GetSchemaResponse, error) {
	if v := req.GetVersion(); v != 0 {
		return nil, fmt.Errorf("unsupported schema version %d", v)
	}
	return &pulumirpc.GetSchemaResponse{Schema: string(p.pulumiSchema)}, nil
}

// Cancel signals the provider to gracefully shut down and abort any ongoing resource operations.
// Operations aborted in this way will return an error (e.g., `Update` and `Create` will either a
// creation error or an initialization error). Since Cancel is advisory and non-blocking, it is up
// to the host to decide how long to wait after Cancel is called before (e.g.)
// hard-closing any gRPC connection.
func (p *cfnProvider) Cancel(context.Context, *pbempty.Empty) (*pbempty.Empty, error) {
	// TODO
	return &pbempty.Empty{}, nil
}

func varsOrEnv(vars map[string]string, key string, env ...string) (string, bool) {
	if val, ok := vars[key]; ok {
		return val, true
	}
	for _, e := range env {
		if val, ok := os.LookupEnv(e); ok {
			return val, true
		}
	}
	return "", false
}

func makeRandom(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	result := make([]rune, length)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}
