package konnect

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/pkg/log"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

type Response interface {
	GetContentType() string
	GetStatusCode() int
	GetRawResponse() *http.Response
}

type Op string

const (
	CreateOp Op = "create"
	UpdateOp Op = "update"
	DeleteOp Op = "delete"
)

func Create[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, logger logr.Logger, cl client.Client, e *T) (*T, error) {
	defer logOpComplete[T, TEnt](logger, time.Now(), CreateOp, e)

	switch ent := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		return e, createControlPlane(ctx, sdk, logger, ent)
	case *configurationv1alpha1.KongService:
		return e, createService(ctx, sdk, logger, cl, ent)
	case *configurationv1alpha1.KongRoute:
		return e, createRoute(ctx, sdk, logger, cl, ent)
	case *configurationv1.KongConsumer:
		return e, createConsumer(ctx, sdk, logger, cl, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return nil, fmt.Errorf("unsupported entity type %T", ent)
	}
}

func Delete[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, logger logr.Logger, cl client.Client, e *T) error {
	defer logOpComplete[T, TEnt](logger, time.Now(), DeleteOp, e)

	switch ent := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		return deleteControlPlane(ctx, sdk, logger, ent)
	case *configurationv1alpha1.KongService:
		return deleteService(ctx, sdk, logger, cl, ent)
	case *configurationv1alpha1.KongRoute:
		return deleteRoute(ctx, sdk, logger, cl, ent)
	case *configurationv1.KongConsumer:
		return deleteConsumer(ctx, sdk, logger, cl, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return fmt.Errorf("unsupported entity type %T", ent)
	}
}

func Update[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, logger logr.Logger, cl client.Client, e *T) (ctrl.Result, error) {
	var (
		ent                = TEnt(e)
		condProgrammed, ok = k8sutils.GetCondition(KonnectEntityProgrammedConditionType, ent)
		now                = time.Now()
		timeFromLastUpdate = time.Since(condProgrammed.LastTransitionTime.Time)
	)
	// If the entity is already programmed and the last update was less than
	// the configured sync period, requeue after the remaining time.
	if ok &&
		condProgrammed.Status == metav1.ConditionTrue &&
		condProgrammed.Reason == KonnectEntityProgrammedReason &&
		condProgrammed.ObservedGeneration == ent.GetObjectMeta().GetGeneration() &&
		timeFromLastUpdate <= configurableSyncPeriod {
		requeueAfter := configurableSyncPeriod - timeFromLastUpdate
		log.Debug(logger, "no need for update, requeueing after configured sync period", e,
			"last_update", condProgrammed.LastTransitionTime.Time,
			"time_from_last_update", timeFromLastUpdate,
			"requeue_after", requeueAfter,
			"requeue_at", now.Add(requeueAfter),
		)
		return ctrl.Result{
			RequeueAfter: requeueAfter,
		}, nil
	}

	defer logOpComplete[T, TEnt](logger, now, UpdateOp, e)

	switch ent := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		return ctrl.Result{}, updateControlPlane(ctx, sdk, logger, ent)
	case *configurationv1alpha1.KongService:
		return ctrl.Result{}, updateService(ctx, sdk, logger, cl, ent)
	case *configurationv1alpha1.KongRoute:
		return ctrl.Result{}, updateRoute(ctx, sdk, logger, cl, ent)
	case *configurationv1.KongConsumer:
		return ctrl.Result{}, updateConsumer(ctx, sdk, logger, cl, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return ctrl.Result{}, fmt.Errorf("unsupported entity type %T", ent)
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
		"konnect_id", e.GetKonnectStatus().GetKonnectID(),
	)
}

// handleResp checks the response from the Konnect API and returns an error if
// the response is not successful.
// It closes the response body.
func handleResp[T SupportedKonnectEntityType](err error, resp Response, op Op) error {
	if err != nil {
		return err
	}
	body := resp.GetRawResponse().Body
	defer body.Close()
	if resp.GetStatusCode() < 200 || resp.GetStatusCode() >= 400 {
		b, err := io.ReadAll(body)
		if err != nil {
			var e T
			return fmt.Errorf(
				"failed to %s %T and failed to read response body: %w",
				op, e, err,
			)
		}
		var e T
		return fmt.Errorf("failed to %s %T: %s", op, e, string(b))
	}
	return nil
}

//nolint:unused
type getLabeler interface {
	GetLabels() map[string]string
}

type SetLabels interface {
	SetKonnectLabels(labels map[string]string)
}

// setKonnectLabels sets the Konnect labels on the object which will be created/updated
// in Konnect.
// TODO: Do we want to copy the k8s labels (or annotations?) to Konnect?
//
//nolint:unused
func setKonnectLabels[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](e TEnt, konnectEntitySpec getLabeler) {
	k8sLabels := k8sLabelsForEntity[T, TEnt](e)

	// Add labels from the Konnect entity spec
	for k, v := range konnectEntitySpec.GetLabels() {
		k8sLabels[k] = v
	}

	if l, ok := any(e).(SetLabels); ok {
		l.SetKonnectLabels(k8sLabels)
	}
}

// k8sLabelsForEntity returns the k8s labels for a Konnect entity.
// Those labels are based on the entity's metadata.
//
//nolint:unused
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
	k8sLabels["k8s-name"] = meta.GetName()
	k8sLabels["k8s-namespace"] = meta.GetNamespace()
	k8sLabels["k8s-managed"] = "true"
	k8sLabels["k8s-generation"] = fmt.Sprintf("%d", meta.GetGeneration())

	return k8sLabels
}
