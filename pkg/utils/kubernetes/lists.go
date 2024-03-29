package kubernetes

import (
	"context"

	admregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListDeploymentsForOwner is a helper function to map a list of Deployments
// by list options and reduce by OwnerReference UID and namespace to efficiently
// list only the objects owned by the provided UID.
func ListDeploymentsForOwner(
	ctx context.Context,
	c client.Client,
	namespace string,
	uid types.UID,
	listOpts ...client.ListOption,
) ([]appsv1.Deployment, error) {
	deploymentList := &appsv1.DeploymentList{}

	err := c.List(
		ctx,
		deploymentList,
		append(
			[]client.ListOption{client.InNamespace(namespace)},
			listOpts...,
		)...,
	)
	if err != nil {
		return nil, err
	}

	deployments := make([]appsv1.Deployment, 0)
	for _, deployment := range deploymentList.Items {
		deployment := deployment
		if IsOwnedByRefUID(&deployment, uid) {
			deployments = append(deployments, deployment)
		}
	}

	return deployments, nil
}

// ListHPAsForOwner is a helper function to map a list of HorizontalPodAutoscalers
// by list options and reduce by OwnerReference UID and namespace to efficiently
// list only the objects owned by the provided UID.
func ListHPAsForOwner(
	ctx context.Context,
	c client.Client,
	namespace string,
	uid types.UID,
	listOpts ...client.ListOption,
) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	hpaList := &autoscalingv2.HorizontalPodAutoscalerList{}

	err := c.List(
		ctx,
		hpaList,
		append(
			[]client.ListOption{client.InNamespace(namespace)},
			listOpts...,
		)...,
	)
	if err != nil {
		return nil, err
	}

	hpas := make([]autoscalingv2.HorizontalPodAutoscaler, 0)
	for _, hpa := range hpaList.Items {
		hpa := hpa
		if IsOwnedByRefUID(&hpa, uid) {
			hpas = append(hpas, hpa)
		}
	}

	return hpas, nil
}

// ListServicesForOwner is a helper function to map a list of Services
// by list options and reduce by OwnerReference UID and namespace to efficiently
// list only the objects owned by the provided UID.
func ListServicesForOwner(
	ctx context.Context,
	c client.Client,
	namespace string,
	uid types.UID,
	listOpts ...client.ListOption,
) ([]corev1.Service, error) {
	serviceList := &corev1.ServiceList{}

	err := c.List(
		ctx,
		serviceList,
		append(
			[]client.ListOption{client.InNamespace(namespace)},
			listOpts...,
		)...,
	)
	if err != nil {
		return nil, err
	}

	services := make([]corev1.Service, 0)
	for _, service := range serviceList.Items {
		service := service
		if IsOwnedByRefUID(&service, uid) {
			services = append(services, service)
		}
	}

	return services, nil
}

// ListServiceAccountsForOwner is a helper function to map a list of ServiceAccounts
// by list options and reduce by OwnerReference UID and namespace to efficiently
// list only the objects owned by the provided UID.
func ListServiceAccountsForOwner(
	ctx context.Context,
	c client.Client,
	namespace string,
	uid types.UID,
	listOpts ...client.ListOption,
) ([]corev1.ServiceAccount, error) {
	serviceAccountList := &corev1.ServiceAccountList{}

	err := c.List(
		ctx,
		serviceAccountList,
		append(
			[]client.ListOption{client.InNamespace(namespace)},
			listOpts...,
		)...,
	)
	if err != nil {
		return nil, err
	}

	serviceAccounts := make([]corev1.ServiceAccount, 0)
	for _, serviceAccount := range serviceAccountList.Items {
		serviceAccount := serviceAccount
		if IsOwnedByRefUID(&serviceAccount, uid) {
			serviceAccounts = append(serviceAccounts, serviceAccount)
		}
	}

	return serviceAccounts, nil
}

// ListClusterRolesForOwner is a helper function to map a list of ClusterRoles
// by list options and reduce by OwnerReference UID to efficiently
// list only the objects owned by the provided UID.
func ListClusterRolesForOwner(
	ctx context.Context,
	c client.Client,
	uid types.UID,
	listOpts ...client.ListOption,
) ([]rbacv1.ClusterRole, error) {
	clusterRoleList := &rbacv1.ClusterRoleList{}

	err := c.List(
		ctx,
		clusterRoleList,
		listOpts...,
	)
	if err != nil {
		return nil, err
	}

	clusterRoles := make([]rbacv1.ClusterRole, 0)
	for _, clusterRole := range clusterRoleList.Items {
		clusterRole := clusterRole
		if IsOwnedByRefUID(&clusterRole, uid) {
			clusterRoles = append(clusterRoles, clusterRole)
		}
	}

	return clusterRoles, nil
}

// ListClusterRoleBindingsForOwner is a helper function to map a list of ClusterRoleBindings
// by list options and reduce by OwnerReference UID to efficiently
// list only the objects owned by the provided UID.
func ListClusterRoleBindingsForOwner(
	ctx context.Context,
	c client.Client,
	uid types.UID,
	listOpts ...client.ListOption,
) ([]rbacv1.ClusterRoleBinding, error) {
	clusterRoleBindingList := &rbacv1.ClusterRoleBindingList{}

	err := c.List(
		ctx,
		clusterRoleBindingList,
		listOpts...,
	)
	if err != nil {
		return nil, err
	}

	clusterRoleBindings := make([]rbacv1.ClusterRoleBinding, 0)
	for _, clusterRoleBinding := range clusterRoleBindingList.Items {
		clusterRoleBinding := clusterRoleBinding
		if IsOwnedByRefUID(&clusterRoleBinding, uid) {
			clusterRoleBindings = append(clusterRoleBindings, clusterRoleBinding)
		}
	}

	return clusterRoleBindings, nil
}

// ListSecretsForOwner is a helper function to map a list of Secrets
// by list options and reduce by OwnerReference UID to efficiently
// list only the objects owned by the provided UID.
func ListSecretsForOwner(ctx context.Context,
	c client.Client,
	uid types.UID,
	listOpts ...client.ListOption,
) ([]corev1.Secret, error) {
	secretList := &corev1.SecretList{}

	err := c.List(
		ctx,
		secretList,
		listOpts...,
	)
	if err != nil {
		return nil, err
	}

	secrets := make([]corev1.Secret, 0)
	for _, secret := range secretList.Items {
		secret := secret
		if IsOwnedByRefUID(&secret, uid) {
			secrets = append(secrets, secret)
		}
	}

	return secrets, nil
}

// ListValidatingWebhookConfigurationsForOwner is a helper function to map a list of ValidatingWebhookConfiguration
// by list options and reduce by OwnerReference UID to efficiently list only the objects owned by the provided UID.
func ListValidatingWebhookConfigurationsForOwner(
	ctx context.Context,
	c client.Client,
	uid types.UID,
	listOpts ...client.ListOption,
) ([]admregv1.ValidatingWebhookConfiguration, error) {
	cfgList := &admregv1.ValidatingWebhookConfigurationList{}
	err := c.List(
		ctx,
		cfgList,
		listOpts...,
	)
	if err != nil {
		return nil, err
	}

	var webhookConfigurations []admregv1.ValidatingWebhookConfiguration
	for _, cfg := range cfgList.Items {
		cfg := cfg
		if IsOwnedByRefUID(&cfg, uid) {
			webhookConfigurations = append(webhookConfigurations, cfg)
		}
	}

	return webhookConfigurations, nil
}
