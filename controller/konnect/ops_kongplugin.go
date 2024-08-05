package konnect

import (
	"context"
	"encoding/json"
	"errors"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/go-logr/logr"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/pkg/op"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

func kongPluginToCreatePlugin(
	plugin *configurationv1.KongPlugin,
	kongService *configurationv1alpha1.KongService,
) (*sdkkonnectgocomp.PluginInput, error) {
	pluginConfig := map[string]any{}
	if err := json.Unmarshal(plugin.Config.Raw, &pluginConfig); err != nil {
		return nil, err
	}

	return &sdkkonnectgocomp.PluginInput{
		Name:    lo.ToPtr(plugin.PluginName),
		Config:  pluginConfig,
		Enabled: lo.ToPtr(!plugin.Disabled),

		Service: &sdkkonnectgocomp.PluginService{
			ID: lo.ToPtr(kongService.Status.Konnect.ID),
		},
	}, nil
}

func upsertPlugin(ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	plugin *sdkkonnectgocomp.PluginInput,
	controlplaneID string,
	pluginBinding *configurationv1alpha1.KongPluginBinding,
) (op.Result, error) {
	var (
		pluginID string
		resp     Response
		respErr  error
		result   op.Result
	)
	if pluginBinding.Status.Konnect.KonnectEntityStatus.ID != "" {
		pluginID = pluginBinding.Status.Konnect.KonnectEntityStatus.ID
	}

	if pluginID != "" {
		upsertResp, err := sdk.Plugins.UpsertPlugin(ctx, operations.UpsertPluginRequest{
			ControlPlaneID: controlplaneID,
			PluginID:       pluginID,
			Plugin:         *plugin,
		})
		if err == nil && upsertResp.Plugin != nil && upsertResp.Plugin.ID != nil {
			pluginID = *upsertResp.Plugin.ID
		}
		respErr = err
		resp = upsertResp
		result = op.Created
	} else {
		// TODO(mlavacca): figure out how to get rid of this CreatePlugin call and use UpsertPlugin only.
		createResp, err := sdk.Plugins.CreatePlugin(ctx, controlplaneID, *plugin)
		if err == nil && createResp.Plugin != nil && createResp.Plugin.ID != nil {
			pluginID = *createResp.Plugin.ID
		}
		respErr = err
		resp = createResp
		result = op.Updated
	}

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errHandled := handleResp[configurationv1alpha1.KongPluginBinding](respErr, resp, CreateOp); errHandled != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				errHandled.Error(),
				pluginBinding.GetGeneration(),
			),
			pluginBinding,
		)
		return op.Noop, errHandled
	}

	pluginBinding.GetKonnectStatus().SetKonnectID(pluginID)
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			pluginBinding.GetGeneration(),
		),
		pluginBinding,
	)

	return result, nil
}

func deletePlugin(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	pluginID string,
	controlplaneID string,
	pluginBinding *configurationv1alpha1.KongPluginBinding,
) (op.Result, error) {
	if pluginID == "" {
		return op.Noop, nil
	}

	resp, err := sdk.Plugins.DeletePlugin(ctx, controlplaneID, pluginID)
	if errHandled := handleResp[configurationv1alpha1.KongRoute](err, resp, DeleteOp); errHandled != nil {
		var sdkError *sdkerrors.SDKError
		if errors.As(errHandled, &sdkError) && sdkError.StatusCode == 404 {
			logger.Info("entity not found in Konnect, skipping delete",
				"op", DeleteOp, "type", pluginBinding.GetTypeName(), "id", pluginID,
			)
			return op.Noop, nil
		}
		return op.Noop, FailedKonnectOpError[configurationv1alpha1.KongRoute]{
			Op:  DeleteOp,
			Err: errHandled,
		}
	}

	return op.Deleted, nil
}
