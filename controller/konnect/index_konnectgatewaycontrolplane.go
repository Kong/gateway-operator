package konnect

import (
	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	"sigs.k8s.io/controller-runtime/pkg/client"

	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

const (
	// IndexFieldKonnectGatewayControlPlaneGroupOnMembers is the index field for KonnectGatewayControlPlane -> its members.
	IndexFieldKonnectGatewayControlPlaneGroupOnMembers = "konnectGatewayControlPlaneGroupMembers"

	// IndexFieldKonnectGatewayControlPlaneOnAPIAuthConfiguration is the index field for KonnectGatewayControlPlane -> APIAuthConfiguration.
	IndexFieldKonnectGatewayControlPlaneOnAPIAuthConfiguration = "konnectGatewayControlPlaneAPIAuthConfigurationRef"

	// IndexFieldKonnectGatewayControlPlaneOnKonnectID is the index field for KonnectGatewayControlPlane -> KonnectID.
	IndexFieldKonnectGatewayControlPlaneOnKonnectID = "konnectGatewayControlPlaneKonnectID"
)

// IndexOptionsForKonnectGatewayControlPlane returns required Index options for KonnectGatewayControlPlane reconciler.
func IndexOptionsForKonnectGatewayControlPlane() []ReconciliationIndexOption {
	return []ReconciliationIndexOption{
		{
			IndexObject:  &konnectv1alpha1.KonnectGatewayControlPlane{},
			IndexField:   IndexFieldKonnectGatewayControlPlaneGroupOnMembers,
			ExtractValue: konnectGatewayControlPlaneGroupMembers,
		},
		{
			IndexObject:  &konnectv1alpha1.KonnectGatewayControlPlane{},
			IndexField:   IndexFieldKonnectGatewayControlPlaneOnAPIAuthConfiguration,
			ExtractValue: konnectGatewayControlPlaneAPIAuthConfigurationRef,
		},
		{
			IndexObject:  &konnectv1alpha1.KonnectGatewayControlPlane{},
			IndexField:   IndexFieldKonnectGatewayControlPlaneOnKonnectID,
			ExtractValue: konnectGatewayControlPlaneKonnectID,
		},
	}
}

func konnectGatewayControlPlaneGroupMembers(object client.Object) []string {
	cp, ok := object.(*konnectv1alpha1.KonnectGatewayControlPlane)
	if !ok {
		return nil
	}
	clusterType := cp.Spec.ClusterType
	if clusterType == nil {
		return nil
	}

	if string(*clusterType) != string(sdkkonnectcomp.CreateControlPlaneRequestClusterTypeClusterTypeControlPlaneGroup) {
		return nil
	}

	ret := make([]string, 0, len(cp.Spec.Members))
	for _, member := range cp.Spec.Members {
		ret = append(ret, member.Name)
	}

	return ret
}

func konnectGatewayControlPlaneAPIAuthConfigurationRef(object client.Object) []string {
	cp, ok := object.(*konnectv1alpha1.KonnectGatewayControlPlane)
	if !ok {
		return nil
	}

	return []string{cp.Spec.KonnectConfiguration.APIAuthConfigurationRef.Name}
}

func konnectGatewayControlPlaneKonnectID(object client.Object) []string {
	cp, ok := object.(*konnectv1alpha1.KonnectGatewayControlPlane)
	if !ok {
		return nil
	}

	if konnectID := cp.GetKonnectStatus().GetKonnectID(); konnectID != "" {
		return []string{konnectID}
	}
	return nil
}
