package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/kong/gateway-operator/apis/v1alpha1"
	operatorv1beta1 "github.com/kong/gateway-operator/apis/v1beta1"
	k8sutils "github.com/kong/gateway-operator/internal/utils/kubernetes"
)

func TestMarkAsProvisioned(t *testing.T) {
	t.Run("dataplane", func(t *testing.T) {
		createDataPlane := func() *operatorv1beta1.DataPlane {
			return &operatorv1beta1.DataPlane{}
		}

		testCases := []struct {
			name              string
			dataplane         func() *operatorv1beta1.DataPlane
			expectedCondition metav1.Condition
		}{
			{
				name: "DataPlane gets a Provisioned Condition with Status True",
				dataplane: func() *operatorv1beta1.DataPlane {
					return createDataPlane()
				},
				expectedCondition: metav1.Condition{
					Type:    string(DataPlaneConditionTypeProvisioned),
					Reason:  string(DataPlaneConditionReasonPodsReady),
					Message: "pods for all Deployments are ready",
					Status:  metav1.ConditionTrue,
				},
			},
			{
				name: "DataPlane gets a Provisioned Condition with Status True and correct ObservedGeneration",
				dataplane: func() *operatorv1beta1.DataPlane {
					dp := createDataPlane()
					dp.Generation = 3
					return dp
				},
				expectedCondition: metav1.Condition{
					Type:               string(DataPlaneConditionTypeProvisioned),
					Reason:             string(DataPlaneConditionReasonPodsReady),
					Message:            "pods for all Deployments are ready",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 3,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				dp := tc.dataplane()
				markAsProvisioned(dp)
				cond, ok := k8sutils.GetCondition(k8sutils.ConditionType(tc.expectedCondition.Type), dp)
				require.True(t, ok)
				assert.Equal(t, cond.Reason, tc.expectedCondition.Reason)
				assert.Equal(t, cond.Status, tc.expectedCondition.Status)
				assert.Equal(t, cond.Message, tc.expectedCondition.Message)
				assert.Equal(t, cond.ObservedGeneration, tc.expectedCondition.ObservedGeneration)
			})
		}
	})

	t.Run("controlplane", func(t *testing.T) {
		createControlPlane := func() *operatorv1alpha1.ControlPlane {
			return &operatorv1alpha1.ControlPlane{}
		}

		testCases := []struct {
			name              string
			controlplane      func() *operatorv1alpha1.ControlPlane
			expectedCondition metav1.Condition
		}{
			{
				name: "ControlPlane gets a Provisioned Condition with Status True",
				controlplane: func() *operatorv1alpha1.ControlPlane {
					return createControlPlane()
				},
				expectedCondition: metav1.Condition{
					Type:    string(ControlPlaneConditionTypeProvisioned),
					Reason:  string(ControlPlaneConditionReasonPodsReady),
					Message: "pods for all Deployments are ready",
					Status:  metav1.ConditionTrue,
				},
			},
			{
				name: "ControlPlane gets a Provisioned Condition with Status True and correct ObservedGeneration",
				controlplane: func() *operatorv1alpha1.ControlPlane {
					cp := createControlPlane()
					cp.Generation = 3
					return cp
				},
				expectedCondition: metav1.Condition{
					Type:               string(ControlPlaneConditionTypeProvisioned),
					Reason:             string(ControlPlaneConditionReasonPodsReady),
					Message:            "pods for all Deployments are ready",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 3,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				dp := tc.controlplane()
				markAsProvisioned(dp)
				cond, ok := k8sutils.GetCondition(k8sutils.ConditionType(tc.expectedCondition.Type), dp)
				require.True(t, ok)
				assert.Equal(t, cond.Reason, tc.expectedCondition.Reason)
				assert.Equal(t, cond.Status, tc.expectedCondition.Status)
				assert.Equal(t, cond.Message, tc.expectedCondition.Message)
				assert.Equal(t, cond.ObservedGeneration, tc.expectedCondition.ObservedGeneration)
			})
		}
	})
}
