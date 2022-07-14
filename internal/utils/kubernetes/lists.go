package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListDeploymentsForOwner is a helper function to map a list of Deployments
// by label and reduce by OwnerReference UID and namespace to efficiently list
// only the objects owned by the provided UID.
func ListDeploymentsForOwner(
	ctx context.Context,
	c client.Client,
	requiredLabel string,
	requiredValue string,
	namespace string,
	uid types.UID,
) ([]appsv1.Deployment, error) {
	listSelector, err := NewListSelectorOption(namespace, requiredLabel, selection.Equals, requiredValue)
	if err != nil {
		return nil, err
	}

	deploymentList := &appsv1.DeploymentList{}
	if err := c.List(ctx, deploymentList, listSelector); err != nil {
		return nil, err
	}

	deployments := make([]appsv1.Deployment, 0)
	for _, deployment := range deploymentList.Items {
		if IsOwnedByRefUID(&deployment.ObjectMeta, uid) {
			deployments = append(deployments, deployment)
		}
	}

	return deployments, nil
}

// ListServicesForOwner is a helper function to map a list of Services
// by label and reduce by OwnerReference UID and namespace to efficiently list
// only the objects owned by the provided UID.
func ListServicesForOwner(
	ctx context.Context,
	c client.Client,
	requiredLabel string,
	requiredValue string,
	namespace string,
	uid types.UID,
) ([]corev1.Service, error) {

	listSelector, err := NewListSelectorOption(namespace, requiredLabel, selection.Equals, requiredValue)
	if err != nil {
		return nil, err
	}

	serviceList := &corev1.ServiceList{}
	if err := c.List(ctx, serviceList, listSelector); err != nil {
		return nil, err
	}

	services := make([]corev1.Service, 0)
	for _, service := range serviceList.Items {
		if IsOwnedByRefUID(&service.ObjectMeta, uid) {
			services = append(services, service)
		}
	}

	return services, nil
}

// NewListSelectorOption is a helper function to create a ListSelectorOption
// from a namespace, label key and label values.
func NewListSelectorOption(
	namespace string,
	labelKey string,
	op selection.Operator,
	labelValues ...string,
) (*client.ListOptions, error) {

	requirement, err := labels.NewRequirement(labelKey, op, labelValues)
	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*requirement)

	return &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: selector,
	}, nil
}
