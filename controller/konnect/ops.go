package konnect

import (
	"context"
	"fmt"
	"time"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kong/gateway-operator/controller/pkg/log"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// Response is the interface for the response from the Konnect API.
type Response interface {
	GetContentType() string
	GetStatusCode() int
}

// Op is the type for the operation type of a Konnect entity.
type Op string

const (
	// CreateOp is the operation type for creating a Konnect entity.
	CreateOp Op = "create"
	// UpdateOp is the operation type for updating a Konnect entity.
	UpdateOp Op = "update"
	// DeleteOp is the operation type for deleting a Konnect entity.
	DeleteOp Op = "delete"
)

// Create creates a Konnect entity.
func Create[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, cl client.Client, e *T) (*T, error) {
	defer logOpComplete[T, TEnt](ctx, time.Now(), CreateOp, e)

	switch ent := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		return e, createControlPlane(ctx, sdk, ent)
	case *configurationv1alpha1.KongService:
		return e, createService(ctx, sdk, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return nil, fmt.Errorf("unsupported entity type %T", ent)
	}
}

// Delete deletes a Konnect entity.
func Delete[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, cl client.Client, e *T) error {
	defer logOpComplete[T, TEnt](ctx, time.Now(), DeleteOp, e)

	switch ent := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		return deleteControlPlane(ctx, sdk, ent)
	case *configurationv1alpha1.KongService:
		return deleteService(ctx, sdk, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return fmt.Errorf("unsupported entity type %T", ent)
	}
}

// Update updates a Konnect entity.
func Update[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, sdk *sdkkonnectgo.SDK, cl client.Client, e *T) (ctrl.Result, error) {
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
		log.Debug(ctrllog.FromContext(ctx),
			"no need for update, requeueing after configured sync period", e,
			"last_update", condProgrammed.LastTransitionTime.Time,
			"time_from_last_update", timeFromLastUpdate,
			"requeue_after", requeueAfter,
			"requeue_at", now.Add(requeueAfter),
		)
		return ctrl.Result{
			RequeueAfter: requeueAfter,
		}, nil
	}

	defer logOpComplete[T, TEnt](ctx, now, UpdateOp, e)

	switch ent := any(e).(type) {
	case *konnectv1alpha1.KonnectControlPlane:
		return ctrl.Result{}, updateControlPlane(ctx, sdk, ent)
	case *configurationv1alpha1.KongService:
		return ctrl.Result{}, updateService(ctx, sdk, cl, ent)

		// ---------------------------------------------------------------------
		// TODO: add other Konnect types

	default:
		return ctrl.Result{}, fmt.Errorf("unsupported entity type %T", ent)
	}
}

func logOpComplete[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](ctx context.Context, start time.Time, op Op, e TEnt) {
	s := e.GetKonnectStatus()
	if s == nil {
		return
	}

	ctrllog.FromContext(ctx).
		Info("operation in Konnect API complete",
			"op", op,
			"duration", time.Since(start),
			"type", entityTypeName[T](),
			"konnect_id", s.GetKonnectID(),
		)
}

// wrapErrIfKonnectOpFailed checks the response from the Konnect API and returns a uniform
// error for all Konnect entities if the operation failed.
func wrapErrIfKonnectOpFailed[
	T SupportedKonnectEntityType,
	TEnt EntityType[T],
](err error, op Op, e TEnt) error {
	if err != nil {
		if e == nil {
			return fmt.Errorf("failed to %s for %T: %w",
				op, e, err,
			)
		}
		return fmt.Errorf("failed to %s for %T %q: %w",
			op, client.ObjectKeyFromObject(e), e, err,
		)
	}
	return nil
}