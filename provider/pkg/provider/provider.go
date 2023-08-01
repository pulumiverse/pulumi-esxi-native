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
	"os"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumiverse/pulumi-esxi-native/provider/pkg/esxi"
)

const (
	logLevel = 9
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

type esxiProvider struct {
	pulumirpc.UnimplementedResourceProviderServer

	host     *provider.HostClient
	name     string
	canceler *cancellationContext

	configured bool
	version    string

	pulumiSchema []byte

	esxi            *esxi.Host
	namingService   *esxi.AutoNamingService
	resourceService *esxi.ResourceService
}

var _ pulumirpc.ResourceProviderServer = (*esxiProvider)(nil)

func newESXiNativeProvider(host *provider.HostClient, name, version string, pulumiSchema []byte) (
	pulumirpc.ResourceProviderServer, error,
) {
	return &esxiProvider{
		host:         host,
		canceler:     makeCancellationContext(),
		name:         name,
		version:      version,
		pulumiSchema: pulumiSchema,
	}, nil
}

// Attach sends the engine address to an already running plugin.
func (p *esxiProvider) Attach(_ context.Context, req *pulumirpc.PluginAttach) (*emptypb.Empty, error) {
	host, err := provider.NewHostClient(req.GetAddress())
	if err != nil {
		return nil, err
	}
	p.host = host
	return &pbempty.Empty{}, nil
}

// Call dynamically executes a method in the provider associated with a component resource.
func (p *esxiProvider) Call(_ context.Context, _ *pulumirpc.CallRequest) (*pulumirpc.CallResponse, error) {
	return nil, status.Error(codes.Unimplemented, "call is not yet implemented")
}

// Construct creates a new component resource.
func (p *esxiProvider) Construct(_ context.Context, _ *pulumirpc.ConstructRequest) (*pulumirpc.ConstructResponse, error) {
	return nil, status.Error(codes.Unimplemented, "construct is not yet implemented")
}

// CheckConfig validates the configuration for this provider.
func (p *esxiProvider) CheckConfig(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.GetNews()}, nil
}

// DiffConfig diffs the configuration for this provider.
func (p *esxiProvider) DiffConfig(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.DiffConfig(%s)", p.name, urn)
	logging.V(logLevel).Infof("%s executing", label)

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

	diffs := make([]string, 0, len(diff.Keys()))
	for _, k := range diff.Keys() {
		diffs = append(diffs, string(k))
	}

	var replaces []string
	return &pulumirpc.DiffResponse{
		Changes:  pulumirpc.DiffResponse_DIFF_SOME,
		Diffs:    diffs,
		Replaces: replaces,
	}, nil
}

// Configure configures the resource provider with "globals" that control its behavior.
func (p *esxiProvider) Configure(_ context.Context, req *pulumirpc.ConfigureRequest) (*pulumirpc.ConfigureResponse, error) {
	vars := req.GetVariables()

	host, hostErr := getConfig(vars, "host", "ESXI_HOST")
	user, userErr := getConfig(vars, "username", "ESXI_USERNAME")
	pass, passErr := getConfig(vars, "password", "ESXI_PASSWORD")
	sshPort, sshPortErr := getConfig(vars, "sshPort", "ESXI_SSH_PORT")
	sslPort, sslPortErr := getConfig(vars, "sslPort", "ESXI_SSL_PORT")
	if len(sshPort) > 0 {
		sshPort = "22"
	}
	if len(sslPort) > 0 {
		sslPort = "443"
	}

	if len(host) > 0 && len(user) > 0 && len(pass) > 0 {
		// If all required values are not present/valid, the client will return an appropriate error.
		esxiHost, err := esxi.NewHost(host, sshPort, sslPort, user, pass)
		if err != nil {
			return nil, err
		}
		p.esxi = esxiHost
	} else {
		errorMessage := "Invalid config."
		for _, errMsg := range []string{hostErr, userErr, passErr, sshPortErr, sslPortErr} {
			if len(errMsg) > 0 {
				errorMessage = fmt.Sprintf("%s\n%s", errorMessage, errMsg)
			}
		}

		return nil, fmt.Errorf(errorMessage)
	}

	p.namingService = esxi.NewAutoNamingService()
	p.resourceService = esxi.NewResourceService()

	p.configured = true

	return &pulumirpc.ConfigureResponse{
		AcceptSecrets: true,
	}, nil
}

// Invoke dynamically executes a built-in function in the provider.
func (p *esxiProvider) Invoke(_ context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	// Unmarshal arguments.
	token := req.GetTok()

	inputs, err := plugin.UnmarshalProperties(req.GetArgs(), plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.Invoke(%s).inputs", p.name, token),
		KeepUnknowns: true,
		KeepSecrets:  true,
	})
	if err != nil {
		return nil, err
	}

	// Process Invoke call.
	result, err := p.resourceService.Invoke(token, inputs, p.esxi)
	if err != nil {
		return nil, err
	}

	res, err := plugin.MarshalProperties(result, plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.Invoke(%s).outputs", p.name, token),
		KeepUnknowns: true,
		KeepSecrets:  true,
	})
	if err != nil {
		return nil, err
	}
	return &pulumirpc.InvokeResponse{Return: res}, nil
}

// StreamInvoke dynamically executes a built-in function in the provider. The result is streamed
// back as a series of messages.
func (p *esxiProvider) StreamInvoke(req *pulumirpc.InvokeRequest, _ pulumirpc.ResourceProvider_StreamInvokeServer) error {
	tok := req.GetTok()
	return fmt.Errorf("unknown StreamInvoke token '%s'", tok)
}

// Check validates that the given property bag is valid for a resource of the given type and returns
// the inputs that should be passed to successive calls to Diff, Create, or Update for this
// resource. As a rule, the provider inputs returned by a call to Check should preserve the original
// representation of the properties as present in the program inputs. Though this rule is not
// required for correctness, violations thereof can negatively impact the end-user experience, as
// the provider inputs are using for detecting and rendering diffs.
func (p *esxiProvider) Check(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Check(%s)", p.name, urn)
	logging.V(logLevel).Infof("%s executing", label)

	resourceToken := string(urn.Type())
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{
		Label: fmt.Sprintf("%s.olds", label), SkipNulls: true,
	})
	if err != nil {
		return nil, err
	}

	// Parse inputs.
	newInputs, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.properties", label),
		KeepUnknowns: true,
		RejectAssets: true,
		KeepSecrets:  true,
	})
	if err != nil {
		return nil, err
	}

	// Parse inputs.
	autoNamingSpec := p.namingService.GetAutoNamingSpec(resourceToken)

	// Filter null properties from the inputs.
	newInputs = filterNullProperties(newInputs)
	if autoNamingSpec != nil {
		// Auto-name fields if not already specified
		val, err := getDefaultName(req.RandomSeed, urn, autoNamingSpec, olds, newInputs)
		if err != nil {
			return nil, err
		}
		newInputs[resource.PropertyKey(autoNamingSpec.PropertyName)] = val
	}
	checkFailures, err := p.resourceService.Validate(resourceToken, newInputs)
	if err != nil {
		return nil, err
	}
	if len(checkFailures) == 0 {
		inputs, err := plugin.MarshalProperties(newInputs, plugin.MarshalOptions{
			Label:        fmt.Sprintf("%s.inputs", label),
			KeepUnknowns: true,
			KeepSecrets:  true,
		})
		if err != nil {
			return nil, err
		}
		return &pulumirpc.CheckResponse{Inputs: inputs}, nil
	}
	return &pulumirpc.CheckResponse{Failures: checkFailures}, nil
}

// Diff checks what impacts a hypothetical update will have on the resource's properties.
func (p *esxiProvider) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Diff(%s)", p.name, urn)
	logging.V(logLevel).Infof("%s executing", label)

	diff, err := p.diffState(req.GetOlds(), req.GetNews(), label)
	if err != nil {
		return nil, err
	}
	if diff == nil {
		return &pulumirpc.DiffResponse{Changes: pulumirpc.DiffResponse_DIFF_NONE}, nil
	}

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_UNKNOWN,
		DeleteBeforeReplace: true,
	}, nil
}

// Create allocates a new instance of the provided resource and returns its unique ID afterward.
func (p *esxiProvider) Create(_ context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Create(%s)", p.name, urn)
	logging.V(logLevel).Infof("%s executing", label)

	// Deserialize RPC inputs.
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.properties", label),
		KeepUnknowns: true,
		RejectAssets: true,
		KeepSecrets:  true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "malformed resource inputs")
	}

	resourceToken := string(urn.Type())
	// Process Create call.
	id, outputs, err := p.resourceService.Create(resourceToken, inputs, p.esxi)
	if err != nil {
		return nil, err
	}

	// Store both outputs and inputs into the state.
	checkpoint, err := plugin.MarshalProperties(
		checkpointObject(inputs, outputs),
		plugin.MarshalOptions{Label: fmt.Sprintf("%s.checkpoint", label), KeepSecrets: true, KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         id,
		Properties: checkpoint,
	}, nil
}

// Read the current live state associated with a resource.
func (p *esxiProvider) Read(_ context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Read(%s)", p.name, urn)
	logging.V(logLevel).Infof("%s executing", label)
	id := req.GetId()

	// Retrieve the old state.
	oldState, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{
		Label: fmt.Sprintf("%s.olds", label), KeepUnknowns: true, SkipNulls: true, KeepSecrets: true,
	})
	if err != nil {
		return nil, err
	}

	// Read the resource state from ESXi.
	resourceToken := string(urn.Type())

	var readInputs resource.PropertyMap

	// Extract old inputs from the `__inputs` field of the old state.
	inputs := parseCheckpointObject(oldState)
	if inputs == nil {
		readInputs = make(resource.PropertyMap)
	} else {
		readInputs = inputs
	}

	// Process Read call.
	id, newState, err := p.resourceService.Read(id, resourceToken, readInputs, p.esxi)
	if err != nil {
		return nil, err
	}

	if inputs == nil {
		inputs = newState
	}

	// Store both outputs and inputs into the state checkpoint.
	checkpoint, err := plugin.MarshalProperties(
		checkpointObject(inputs, newState),
		plugin.MarshalOptions{Label: fmt.Sprintf("%s.checkpoint", label), KeepSecrets: true, KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	// Serialize and return the calculated inputs.
	inputsRecord, err := plugin.MarshalProperties(
		inputs,
		plugin.MarshalOptions{Label: fmt.Sprintf("%s.inputs", label), KeepSecrets: true, KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         id,
		Inputs:     inputsRecord,
		Properties: checkpoint,
	}, nil
}

// Update updates an existing resource with new values.
func (p *esxiProvider) Update(_ context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Update(%s)", p.name, urn)
	logging.V(logLevel).Infof("%s executing", label)

	id := req.GetId()
	resourceToken := string(urn.Type())

	// Read the inputs to persist them into state.
	newInputs, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.newInputs", label),
		KeepUnknowns: true,
		RejectAssets: true,
		KeepSecrets:  true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "diff failed because malformed resource inputs")
	}

	// Process Update call.
	outputs, err := p.resourceService.Update(id, resourceToken, make(resource.PropertyMap), p.esxi)
	if err != nil {
		return nil, err
	}

	// Store both outputs and inputs into the state and return RPC checkpoint.
	checkpoint, err := plugin.MarshalProperties(
		checkpointObject(newInputs, outputs),
		plugin.MarshalOptions{Label: fmt.Sprintf("%s.checkpoint", label), KeepSecrets: true, KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{Properties: checkpoint}, nil
}

// Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed
// to still exist.
func (p *esxiProvider) Delete(_ context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	urn := resource.URN(req.GetUrn())
	label := fmt.Sprintf("%s.Update(%s)", p.name, urn)
	logging.V(logLevel).Infof("%s executing", label)

	resourceToken := string(urn.Type())
	id := req.GetId()

	// Process Read call.
	err := p.resourceService.Delete(resourceToken, id, p.esxi)
	if err != nil {
		return nil, err
	}

	return &pbempty.Empty{}, nil
}

// GetPluginInfo returns generic information about this plugin, like its version.
func (p *esxiProvider) GetPluginInfo(context.Context, *pbempty.Empty) (*pulumirpc.PluginInfo, error) {
	return &pulumirpc.PluginInfo{
		Version: p.version,
	}, nil
}

// GetSchema returns the JSON-serialized schema for the provider.
func (p *esxiProvider) GetSchema(_ context.Context, req *pulumirpc.GetSchemaRequest) (*pulumirpc.GetSchemaResponse, error) {
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
func (p *esxiProvider) Cancel(context.Context, *pbempty.Empty) (*pbempty.Empty, error) {
	p.canceler.cancel()
	return &pbempty.Empty{}, nil
}

// diffState extracts old and new inputs and calculates a diff between them.
func (p *esxiProvider) diffState(olds *structpb.Struct, news *structpb.Struct, label string) (*resource.ObjectDiff, error) {
	oldState, err := plugin.UnmarshalProperties(olds, plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.oldState", label),
		KeepUnknowns: true,
		RejectAssets: true,
		KeepSecrets:  true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "diff failed because malformed resource inputs")
	}

	// Extract old inputs from the `__inputs` field of the old state.
	oldInputs := parseCheckpointObject(oldState)

	newInputs, err := plugin.UnmarshalProperties(news, plugin.MarshalOptions{
		Label:        fmt.Sprintf("%s.newInputs", label),
		KeepUnknowns: true,
		RejectAssets: true,
		KeepSecrets:  true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "diff failed because malformed resource inputs")
	}

	return oldInputs.Diff(newInputs), nil
}

// checkpointObject puts inputs in the `__inputs` field of the state.
func checkpointObject(inputs resource.PropertyMap, outputs resource.PropertyMap) resource.PropertyMap {
	object := outputs
	object["__inputs"] = resource.MakeSecret(resource.NewObjectProperty(inputs))
	return object
}

// parseCheckpointObject returns inputs that are saved in the `__inputs` field of the state.
func parseCheckpointObject(obj resource.PropertyMap) resource.PropertyMap {
	if inputs, ok := obj["__inputs"]; ok {
		return inputs.SecretValue().Element.ObjectValue()
	}

	return nil
}

// filterNullValues removes nested null values from the given property value. If all nested values in the property
// value are null, the property value itself is considered null. This allows for easier integration with CloudFormation
// templates, which use `AWS::NoValue` to remove values from lists, maps, etc. We use `null` to the same effect.
func filterNullValues(v resource.PropertyValue) resource.PropertyValue {
	switch {
	case v.IsArray():
		if len(v.ArrayValue()) == 0 {
			return v
		}

		var result []resource.PropertyValue
		for _, e := range v.ArrayValue() {
			e = filterNullValues(e)
			if !e.IsNull() {
				result = append(result, e)
			}
		}
		return resource.NewArrayProperty(result)
	case v.IsObject():
		return resource.NewObjectProperty(filterNullProperties(v.ObjectValue()))
	default:
		return v
	}
}

// filterNullValues removes nested null values from the given property map. If all nested values in the property
// map are null, the property map itself is considered null. This allows for easier integration with CloudFormation
// templates, which use `AWS::NoValue` to remove values from lists, maps, etc. We use `null` to the same effect.
func filterNullProperties(m resource.PropertyMap) resource.PropertyMap {
	if len(m) == 0 {
		return m
	}

	result := resource.PropertyMap{}
	for k, v := range m {
		e := filterNullValues(v)
		if !e.IsNull() {
			result[k] = e
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func getConfig(vars map[string]string, key, env string) (string, string) {
	if val, ok := vars[fmt.Sprintf("esxi-native:config:%s", key)]; ok {
		return val, ""
	}
	if val, ok := os.LookupEnv(env); ok {
		return val, ""
	}
	return "", fmt.Sprintf("config key 'esxi-native:config:%s', or env var.: '%s', must be provided", key, env)
}
