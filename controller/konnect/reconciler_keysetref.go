package konnect

import (
	"context"
	"fmt"

	"github.com/samber/mo"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kong/gateway-operator/controller/konnect/constraints"
	"github.com/kong/gateway-operator/controller/pkg/patch"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// handleKongKeySetRef handles the KeySetRef for the given entity.
func handleKongKeySetRef[T constraints.SupportedKonnectEntityType, TEnt constraints.EntityType[T]](
	ctx context.Context,
	cl client.Client,
	ent TEnt,
) (ctrl.Result, error) {
	keySetRef, ok := getKeySetRef(ent).Get()
	if !ok {
		if key, ok := any(ent).(*configurationv1alpha1.KongKey); ok {
			// If the entity has a resolved reference, but the spec has changed, we need to adjust the status
			// and transfer the ownership back from the KeySet to the ControlPlane.
			if key.Status.Konnect != nil && key.Status.Konnect.GetKeySetID() != "" {
				old := key.DeepCopyObject().(TEnt)
				// Reset the KeySetID in the status and set the condition to True.
				key.Status.Konnect.KeySetID = ""
				_ = patch.SetStatusWithConditionIfDifferent(ent,
					konnectv1alpha1.KeySetRefValidConditionType,
					metav1.ConditionTrue,
					konnectv1alpha1.KeySetRefReasonValid,
					"KeySetRef is unset",
				)

				// Patch the status
				if _, err := patch.ApplyStatusPatchIfNotEmpty(ctx, cl, ctrllog.FromContext(ctx), ent, old); err != nil {
					if k8serrors.IsConflict(err) {
						return ctrl.Result{Requeue: true}, nil
					}
					return ctrl.Result{}, fmt.Errorf("failed to patch status: %w", err)
				}

				// Transfer the ownership back to the ControlPlane if it's resolved.
				cpRef, hasCPRef := getControlPlaneRef(ent).Get()
				if hasCPRef {
					cp, err := getCPForRef(ctx, cl, cpRef, key.GetNamespace())
					if err != nil {
						return ctrl.Result{}, fmt.Errorf("failed to get ControlPlane: %w", err)
					}
					if res, err := passOwnershipExclusivelyTo(ctx, cl, key, cp); err != nil || !res.IsZero() {
						return res, fmt.Errorf("failed to transfer ownership to ControlPlane: %w", err)
					}
				}
			}
		}
		return ctrl.Result{}, nil
	}

	if keySetRef.Type != configurationv1alpha1.KeySetRefNamespacedRef {
		ctrllog.FromContext(ctx).Error(fmt.Errorf("unsupported KeySet ref type %q", keySetRef.Type), "unsupported KeySet ref type", "entity", ent)
		return ctrl.Result{}, nil
	}

	keySet := configurationv1alpha1.KongKeySet{}
	nn := types.NamespacedName{
		Name:      keySetRef.NamespacedRef.Name,
		Namespace: ent.GetNamespace(),
	}
	if err := cl.Get(ctx, nn, &keySet); err != nil {
		if res, errStatus := patch.StatusWithCondition(
			ctx, cl, ent,
			konnectv1alpha1.KeySetRefValidConditionType,
			metav1.ConditionFalse,
			konnectv1alpha1.KeySetRefReasonInvalid,
			err.Error(),
		); errStatus != nil || !res.IsZero() {
			return res, errStatus
		}

		// If the KongKeySet is not found, we don't want to requeue.
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, ReferencedKongKeySetDoesNotExist{
				Reference: nn,
				Err:       err,
			}
		}

		return ctrl.Result{}, fmt.Errorf("failed getting KongKeySet %s: %w", nn, err)
	}

	// If referenced KongKeySet is being deleted, return an error so that we can remove the entity from Konnect first.
	if delTimestamp := keySet.GetDeletionTimestamp(); !delTimestamp.IsZero() {
		return ctrl.Result{}, ReferencedKongKeySetIsBeingDeleted{
			Reference:         nn,
			DeletionTimestamp: delTimestamp.Time,
		}
	}

	// Verify that the KongKeySet is programmed.
	cond, ok := k8sutils.GetCondition(konnectv1alpha1.KonnectEntityProgrammedConditionType, &keySet)
	if !ok || cond.Status != metav1.ConditionTrue {
		if res, err := patch.StatusWithCondition(
			ctx, cl, ent,
			konnectv1alpha1.KeySetRefValidConditionType,
			metav1.ConditionFalse,
			konnectv1alpha1.KeySetRefReasonInvalid,
			fmt.Sprintf("Referenced KongKeySet %s is not programmed yet", nn),
		); err != nil || !res.IsZero() {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Transfer the ownership of the entity exclusively to the KongKeySet to make sure it will get garbage collected
	// when the KongKeySet is deleted. This is to follow the behavior on the Konnect API that deletes KongKeys associated
	// with a KongKeySet once it's deleted.
	// The ownership needs to be transferred *exclusively* to the KongKeySet because a Kubernetes object gets garbage
	// collected only when all its owner references are removed.
	if res, err := passOwnershipExclusivelyTo(ctx, cl, ent, &keySet); err != nil || !res.IsZero() {
		return res, err
	}

	old := ent.DeepCopyObject().(TEnt)

	// TODO: make this generic.
	// KongKeySet ID is not stored in KonnectEntityStatus because not all entities
	// have a KeySetRef, hence the type constraints in the reconciler can't be used.
	if key, ok := any(ent).(*configurationv1alpha1.KongKey); ok {
		if key.Status.Konnect == nil {
			key.Status.Konnect = &konnectv1alpha1.KonnectEntityStatusWithControlPlaneAndKeySetRef{}
		}
		key.Status.Konnect.KeySetID = keySet.Status.Konnect.GetKonnectID()
	}

	_ = patch.SetStatusWithConditionIfDifferent(ent,
		konnectv1alpha1.KeySetRefValidConditionType,
		metav1.ConditionTrue,
		konnectv1alpha1.KeySetRefReasonValid,
		fmt.Sprintf("Referenced KongKeySet %s programmed", nn),
	)

	_, err := patch.ApplyStatusPatchIfNotEmpty(ctx, cl, ctrllog.FromContext(ctx), ent, old)
	if err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func getKeySetRef[T constraints.SupportedKonnectEntityType, TEnt constraints.EntityType[T]](
	e TEnt,
) mo.Option[configurationv1alpha1.KeySetRef] {
	switch e := any(e).(type) {
	case *configurationv1alpha1.KongKey:
		if e.Spec.KeySetRef == nil {
			return mo.None[configurationv1alpha1.KeySetRef]()
		}
		return mo.Some(*e.Spec.KeySetRef)
	default:
		return mo.None[configurationv1alpha1.KeySetRef]()
	}
}

// passOwnershipExclusivelyTo transfers the ownership of the entity exclusively to the given owner, removing all other
// owner references.
func passOwnershipExclusivelyTo[T constraints.SupportedKonnectEntityType, TEnt constraints.EntityType[T]](
	ctx context.Context,
	cl client.Client,
	ent TEnt,
	to metav1.Object,
) (ctrl.Result, error) {
	old := ent.DeepCopyObject().(TEnt)

	// Cleanup the old owner references.
	ent.SetOwnerReferences(nil)

	// Set the owner reference.
	if err := controllerutil.SetOwnerReference(to, ent, cl.Scheme(), controllerutil.WithBlockOwnerDeletion(true)); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Patch the spec
	if _, _, err := patch.ApplyPatchIfNotEmpty(ctx, cl, ctrllog.FromContext(ctx), ent, old, true); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to patch: %w", err)
	}

	return ctrl.Result{}, nil
}
