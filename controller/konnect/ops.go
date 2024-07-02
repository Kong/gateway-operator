package konnect

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	"github.com/go-logr/logr"

	configurationv1alpha1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1alpha1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
)

type Response interface {
	GetContentType() string
	GetStatusCode() int
	GetRawResponse() *http.Response
}

type Op string

const (
	GetOp    Op = "get"
	CreateOp Op = "create"
	UpdateOp Op = "update"
	DeleteOp Op = "delete"
)

// func Get[
// 	T SupportedKonnectEntityType,
// ](ctx context.Context, sdk *sdkkonnectgo.SDK, id string) (*T, error) {
// 	var e T
// 	switch ent := any(e).(type) {
// 	case operatorv1alpha1.KonnectControlPlane:
// 		resp, err := sdk.ControlPlanes.Get(ctx, id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if err := handleStatusCode[T](resp, GetOp); err != nil {
// 			return nil, err
// 		}

// 		resp.ControlPlane
// 		ent.Status.KonnectID = *resp.ControlPlane.ID
// 		// TODO: add other types
// 		return e, nil

// 	default:
// 		return nil, fmt.Errorf("unsupported entity type %T", ent)
// 	}

// 	return nil, nil
// }

func Create[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, logger logr.Logger, e *T) (*T, error) {
	defer logOpComplete[T, TEnt](logger, time.Now(), CreateOp, e)

	switch ent := any(e).(type) {
	case *operatorv1alpha1.KonnectControlPlane:
		return e, createControlPlane(ctx, sdk, logger, ent)
	case *configurationv1alpha1.Service:
		return e, createService(ctx, sdk, logger, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return nil, fmt.Errorf("unsupported entity type %T", ent)
	}
}

func Delete[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, logger logr.Logger, e *T) error {
	defer logOpComplete[T, TEnt](logger, time.Now(), CreateOp, e)

	switch ent := any(e).(type) {
	case *operatorv1alpha1.KonnectControlPlane:
		return deleteControlPlane(ctx, sdk, logger, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return fmt.Errorf("unsupported entity type %T", ent)
	}
}

func Update[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, logger logr.Logger, e *T) error {
	defer logOpComplete[T, TEnt](logger, time.Now(), UpdateOp, e)

	switch ent := any(e).(type) {
	case *operatorv1alpha1.KonnectControlPlane:
		return updateControlPlane(ctx, sdk, logger, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return fmt.Errorf("unsupported entity type %T", ent)
	}
}

func logOpComplete[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](logger logr.Logger, start time.Time, op Op, e TEnt) {
	logger.Info("operation in Konnect API complete",
		"op", op,
		"duration", time.Since(start),
		"type", entityTypeName[T](),
		"konnect_id", e.GetStatus().GetKonnectID(),
	)
}

func handleResp[T SupportedKonnectEntityType](err error, resp Response, op Op) error {
	if err != nil {
		return err
	}
	if resp.GetStatusCode() < 200 || resp.GetStatusCode() >= 400 {
		b, err := io.ReadAll(resp.GetRawResponse().Body)
		if err != nil {
			var e T
			return fmt.Errorf(
				"failed to %s %T and failed to read response body: %v",
				op, e, err,
			)
		}
		var e T
		return fmt.Errorf("failed to %s %T: %s", op, e, string(b))
	}
	return nil
}

type getLabeler interface {
	GetLabels() map[string]string
}

// setKonnectLabels sets the Konnect labels on the object which will be created/updated
// in Konnect.
// TODO: Do we want to copy the k8s labels (or annotations?) to Konnect?
func setKonnectLabels[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](e TEnt, konnectEntitySpec getLabeler) {
	k8sLabels := k8sLabelsForEntity[T, TEnt](e)

	// Add labels from the Konnect entity spec
	for k, v := range konnectEntitySpec.GetLabels() {
		k8sLabels[k] = v
	}

	e.SetKonnectLabels(k8sLabels)
}

// k8sLabelsForEntity returns the k8s labels for a Konnect entity.
// Those labels are based on the entity's metadata.
func k8sLabelsForEntity[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](e TEnt) map[string]string {
	meta := e.GetObjectMeta()
	k8sLabels := meta.GetLabels()
	if k8sLabels == nil {
		k8sLabels = make(map[string]string)
	}
	k8sLabels["k8s-uid"] = string(meta.GetUID())
	k8sLabels["k8s-name"] = string(meta.GetName())
	k8sLabels["k8s-namespace"] = string(meta.GetNamespace())
	k8sLabels["k8s-managed"] = "true"
	k8sLabels["k8s-generation"] = fmt.Sprintf("%d", meta.GetGeneration())

	return k8sLabels
}
