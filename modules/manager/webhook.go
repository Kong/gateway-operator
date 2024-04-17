/*
Copyright 2022 Kong Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/kong/gateway-operator/modules/admission"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	k8sresources "github.com/kong/gateway-operator/pkg/utils/kubernetes/resources"
)

const (
	defaultWebhookCertDir = "/tmp/k8s-webhook-server/serving-certs"

	defaultsecretPollInterval = 2 * time.Second
	defaultsecretPollTimeout  = 60 * time.Second
)

type webhookManager struct {
	client client.Client
	mgr    ctrl.Manager
	logger logr.Logger
	cfg    *Config
	server webhook.Server

	setupControllers        SetupControllersFunc
	admissionRequestHandler AdmissionRequestHandlerFunc
}

// AdmissionRequestHandlerFunc is a function that returns an implementation of admission.RequestHandler,
// (validation webhook) it's passed to Run function and called later.
type AdmissionRequestHandlerFunc func(c client.Client, l logr.Logger) *admission.RequestHandler

// PrepareWebhookServerWithControllers creates a webhook server and adds it to the controller manager.
// Because the controller runtime 0.14.x doed not allow adding readiness probe after manager starts,
// We need to create webhook server and add it to manager before manager starts.
// https://github.com/Kong/gateway-operator/issues/611
func (m *webhookManager) PrepareWebhookServerWithControllers(
	ctx context.Context,
	setupControllers SetupControllersFunc,
	newAdmissionRequestHandler AdmissionRequestHandlerFunc,
) error {
	if m.cfg.ControllerNamespace == "" {
		return errors.New("controllerNamespace must be set")
	}
	if m.cfg.WebhookCertDir == "" {
		return errors.New("webhookCertDir must be set")
	}
	if m.cfg.WebhookPort == 0 {
		return errors.New("webhookPort must be set")
	}
	m.setupControllers = setupControllers

	// create and start a new webhook server
	hookServer := webhook.NewServer(webhook.Options{
		CertDir: m.cfg.WebhookCertDir,
		Port:    m.cfg.WebhookPort,
	})

	// add readyz check for checking connection to webhook server
	// to make the controller to be marked as ready after webhook started.
	if err := m.mgr.AddReadyzCheck("readyz", hookServer.StartedChecker()); err != nil {
		return fmt.Errorf("failed to add readiness probe for webhook: %w", err)
	}

	m.server = hookServer
	m.admissionRequestHandler = newAdmissionRequestHandler

	return nil
}

// Start starts the webhook server and the controllers.
func (m *webhookManager) Start(ctx context.Context) error {
	m.logger.Info("starting webhook manager")

	// create the webhook resources (if they already exist, it is no-op)
	if err := m.createWebhookResources(ctx); err != nil {
		return err
	}

	certSecret := &corev1.Secret{}
	// check if the certificate secret already exists
	if err := m.client.Get(ctx, types.NamespacedName{Namespace: m.cfg.ControllerNamespace, Name: consts.WebhookCertificateConfigSecretName}, certSecret); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		// no certificate secret found, create all the resources needed to produce it (if they already exist, it is no-op)
		if err := m.createCertificateConfigResources(ctx); err != nil {
			return err
		}

		// wait for the certificate to be created
		certSecret, err = m.waitForWebhookCertificate(ctx, defaultsecretPollTimeout, defaultsecretPollInterval)
		if err != nil {
			return err
		}
	}

	// write the webhook certificate files on the filesystem
	{
		p := path.Join(m.cfg.WebhookCertDir, caCertFilename)
		if err := os.WriteFile(p, certSecret.Data[consts.CAFieldSecret], os.ModePerm); err != nil {
			return fmt.Errorf("failed writing CA to %s: %w", p, err)
		}
	}
	{
		p := path.Join(m.cfg.WebhookCertDir, tlsCertFilename)
		if err := os.WriteFile(p, certSecret.Data[consts.CertFieldSecret], os.ModePerm); err != nil {
			return fmt.Errorf("failed writing certificate to %s: %w", p, err)
		}
	}
	{
		p := path.Join(m.cfg.WebhookCertDir, tlsKeyFilename)
		if err := os.WriteFile(p, certSecret.Data[consts.KeyFieldSecret], os.ModePerm); err != nil {
			return fmt.Errorf("failed writing key to %s: %w", p, err)
		}
	}

	handler := m.admissionRequestHandler(m.mgr.GetClient(), m.logger)
	m.server.Register("/validate", handler)
	if err := m.mgr.Add(m.server); err != nil {
		return err
	}

	// load the Gateway API controllers and start them only after the webhook is in place
	controllers, err := m.setupControllers(m.mgr, m.cfg)
	if err != nil {
		return err
	}

	for _, c := range controllers {
		if err := c.MaybeSetupWithManager(m.mgr); err != nil {
			return fmt.Errorf("unable to create controller %q: %w", c.Name(), err)
		}
	}

	return nil
}

// createCertificateConfigResources create all the resources needed by the CertificateConfig jobs
func (m *webhookManager) createCertificateConfigResources(ctx context.Context) error {
	// create the certificateConfig ServiceAccount
	serviceAccount := k8sresources.GenerateNewServiceAccountForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.client.Create(ctx, serviceAccount); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// create the certificateConfig ClusterRole
	clusterRole := k8sresources.GenerateNewClusterRoleForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.setNamespaceAsOwner(ctx, clusterRole); err != nil {
		return err
	}
	if err := m.client.Create(ctx, clusterRole); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// create the certificateConfig ClusterRoleBinding
	clusterRoleBinding := k8sresources.GenerateNewClusterRoleBindingForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.setNamespaceAsOwner(ctx, clusterRoleBinding); err != nil {
		return err
	}
	if err := m.client.Create(ctx, clusterRoleBinding); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// create the certificateConfig Role
	role := k8sresources.GenerateNewRoleForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.client.Create(ctx, role); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// create the certificateConfig RoleBinding
	roleBinding := k8sresources.GenerateNewRoleBindingForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.client.Create(ctx, roleBinding); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// create the certificateConfig jobs
	if err := m.createCertificateConfigJobs(ctx); err != nil {
		return err
	}

	return nil
}

func (m *webhookManager) createWebhookResources(ctx context.Context) error {
	// create the operator ValidatingWebhookConfiguration
	validatingWebhookConfiguration := k8sresources.
		NewValidatingWebhookConfigurationBuilder(consts.WebhookName).
		WithClientConfigKubernetesService(
			types.NamespacedName{
				Name:      consts.WebhookServiceName,
				Namespace: m.cfg.ControllerNamespace,
			},
		).
		Build()
	if err := m.setNamespaceAsOwner(ctx, validatingWebhookConfiguration); err != nil {
		return err
	}
	if err := m.client.Create(ctx, validatingWebhookConfiguration); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// create the Service needed to expose the operator Webhook
	webhookService := k8sresources.GenerateNewServiceForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookServiceName)
	if err := m.client.Create(ctx, webhookService); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (m *webhookManager) createCertificateConfigJobs(ctx context.Context) error {
	jobCertificateConfigImage := consts.WebhookCertificateConfigBaseImage
	if relatedJobImage := os.Getenv("RELATED_IMAGE_CERTIFICATE_CONFIG"); relatedJobImage != "" {
		// RELATED_IMAGE_CERTIFICATE_CONFIG is set by the operator-sdk when building the operator bundle.
		// https://github.com/Kong/gateway-operator-archive/issues/261
		jobCertificateConfigImage = relatedJobImage
	}
	job := k8sresources.GenerateNewWebhookCertificateConfigJob(
		m.cfg.ControllerNamespace,
		consts.WebhookCertificateConfigName,
		jobCertificateConfigImage,
		consts.WebhookCertificateConfigSecretName,
		consts.WebhookName,
	)

	if err := m.client.Create(ctx, job); err != nil {
		return err
	}

	return nil
}

func (m *webhookManager) cleanup(ctx context.Context) error {
	m.logger.Info("cleaning up webhook resources")

	if err := m.cleanupCertificateConfigResources(ctx); err != nil {
		return err
	}

	return m.cleanupWebhookResources(ctx)
}

func (m *webhookManager) cleanupWebhookResources(ctx context.Context) error {
	// delete the operator ValidatingWebhookConfiguration
	validatingWebhookConfiguration := k8sresources.
		NewValidatingWebhookConfigurationBuilder(consts.WebhookName).
		WithClientConfigKubernetesService(
			types.NamespacedName{
				Name:      consts.WebhookServiceName,
				Namespace: m.cfg.ControllerNamespace,
			},
		).
		Build()
	if err := m.client.Delete(ctx, validatingWebhookConfiguration); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	// delete the Service needed to expose the operator Webhook
	webhookService := k8sresources.GenerateNewServiceForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookServiceName)
	if err := m.client.Delete(ctx, webhookService); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	certSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      consts.WebhookCertificateConfigSecretName,
			Namespace: m.cfg.ControllerNamespace,
		},
	}
	if err := m.client.Delete(ctx, certSecret); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (m *webhookManager) cleanupCertificateConfigResources(ctx context.Context) error {
	// delete the certificateConfig ServiceAccount
	serviceAccount := k8sresources.GenerateNewServiceAccountForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.client.Delete(ctx, serviceAccount); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	// delete the certificateConfig ClusterRole
	clusterRole := k8sresources.GenerateNewClusterRoleForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.client.Delete(ctx, clusterRole); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	// delete the certificateConfig ClusterRoleBinding
	clusterRoleBinding := k8sresources.GenerateNewClusterRoleBindingForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.client.Delete(ctx, clusterRoleBinding); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	// delete the certificateConfig Role
	role := k8sresources.GenerateNewRoleForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.client.Delete(ctx, role); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	// delete the certificateConfig RoleBinding
	roleBinding := k8sresources.GenerateNewRoleBindingForCertificateConfig(m.cfg.ControllerNamespace, consts.WebhookCertificateConfigName, consts.WebhookCertificateConfigLabelvalue)
	if err := m.client.Delete(ctx, roleBinding); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// waitForWebhookCertificate polls the API server at a specific interval until the webhook certificate
// secret is created. If the timer expires, it returns an error. Otherwise, the Secret is returned.
func (m *webhookManager) waitForWebhookCertificate(ctx context.Context, pollTimeout time.Duration, pollInterval time.Duration) (*corev1.Secret, error) {
	ticker := time.NewTicker(pollInterval)
	quit := make(chan struct{})
	errChan := make(chan error, 1)
	certificateSecret := &corev1.Secret{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				err := m.client.Get(ctx, types.NamespacedName{Namespace: m.cfg.ControllerNamespace, Name: consts.WebhookCertificateConfigSecretName}, certificateSecret)
				if err != nil {
					if !k8serrors.IsNotFound(err) {
						errChan <- err
						return
					}
					continue
				}
				return
			case <-quit:
				ticker.Stop()
				errChan <- fmt.Errorf("timeout for creating webhook certificate expired")
				return
			}
		}
	}()
	time.AfterFunc(pollTimeout, func() {
		close(quit)
	})
	wg.Wait()
	ticker.Stop()
	select {
	case err := <-errChan:
		return nil, err
	default:
		return certificateSecret, nil
	}
}

// setNamespaceAsOwner sets the namespace as ownerReference for the given objects.
// This is needed by the operator-related cluster-wide resources that have to be
// collected when the namespace in which the operator lives is deleted
// (e.g., when 'kubectl kustomize config/default | kubectl delete -f -' is executed).
func (m *webhookManager) setNamespaceAsOwner(ctx context.Context, object client.Object) error {
	namespace := &corev1.Namespace{}
	if err := m.client.Get(ctx, types.NamespacedName{Name: m.cfg.ControllerNamespace}, namespace); err != nil {
		return err
	}
	k8sutils.SetOwnerForObject(object, namespace)
	return nil
}
