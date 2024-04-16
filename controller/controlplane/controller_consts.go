package controlplane

// -----------------------------------------------------------------------------
// ControlPlane - Finalizers
// -----------------------------------------------------------------------------

// ControlPlaneFinalizer defines finalizers added by controlplane controller.
type ControlPlaneFinalizer string

const (
	// ControlPlaneFinalizerCleanupClusterRole is the finalizer to cleanup clusterroles owned by controlplane on deleting.
	ControlPlaneFinalizerCleanupClusterRole ControlPlaneFinalizer = "gateway-operator.konghq.com/cleanup-clusterrole"
	// ControlPlaneFinalizerCleanupClusterRoleBinding is the finalizer to cleanup clusterrolebindings owned by controlplane on deleting.
	ControlPlaneFinalizerCleanupClusterRoleBinding ControlPlaneFinalizer = "gateway-operator.konghq.com/cleanup-clusterrolebinding"
	// ControlPlaneFinalizerCleanupValidatingWebhookConfiguration is the finalizer to cleanup validatingwebhookconfigurations owned by controlplane on deleting.
	ControlPlaneFinalizerCleanupValidatingWebhookConfiguration ControlPlaneFinalizer = "gateway-operator.konghq.com/cleanup-validatingwebhookconfiguration"
)
