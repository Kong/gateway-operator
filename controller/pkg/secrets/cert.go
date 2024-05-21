package secrets

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cloudflare/cfssl/config"
	cflog "github.com/cloudflare/cfssl/log"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
	"github.com/go-logr/logr"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	"github.com/kong/gateway-operator/controller/pkg/dataplane"
	"github.com/kong/gateway-operator/controller/pkg/op"
	"github.com/kong/gateway-operator/modules/manager/logging"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
	k8sreduce "github.com/kong/gateway-operator/pkg/utils/kubernetes/reduce"
	k8sresources "github.com/kong/gateway-operator/pkg/utils/kubernetes/resources"
)

// -----------------------------------------------------------------------------
// Private Functions - Certificate management
// -----------------------------------------------------------------------------

// cfssl uses its own internal logger which will yeet unformatted messages to stderr unless overidden
type loggerShim struct {
	logger logr.Logger
}

// Debug logs on debug level.
func (l loggerShim) Debug(msg string) { l.logger.V(logging.DebugLevel.Value()).Info(msg) }

// Info logs on info level.
func (l loggerShim) Info(msg string) { l.logger.V(logging.DebugLevel.Value()).Info(msg) }

// Warning logs on warning level.
func (l loggerShim) Warning(msg string) { l.logger.V(logging.InfoLevel.Value()).Info(msg) }

// Err logs on error level.
func (l loggerShim) Err(msg string) { l.logger.V(logging.InfoLevel.Value()).Info(msg) }

// Crit logs on critical level.
func (l loggerShim) Crit(msg string) { l.logger.V(logging.InfoLevel.Value()).Info(msg) }

// Emerg logs on emergency level.
func (l loggerShim) Emerg(msg string) { l.logger.V(logging.InfoLevel.Value()).Info(msg) }

var caLoggerInit sync.Once

func setCALogger(logger logr.Logger) {
	caLoggerInit.Do(func() {
		cflog.SetLogger(loggerShim{logger: logger})
	})
}

/*
Adapted from the Kubernetes CFSSL signer:
https://github.com/kubernetes/kubernetes/blob/v1.16.15/pkg/controller/certificates/signer/cfssl_signer.go
Modified to handle requests entirely in memory instead of via a controller watching for CertificateSigningRequests
in the API.


Copyright 2016 The Kubernetes Authors.
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

// signCertificate takes a CertificateSigningRequest and a TLS Secret and returns a PEM x.509 certificate
// signed by the certificate in the Secret.
func signCertificate(csr certificatesv1.CertificateSigningRequest, ca *corev1.Secret) ([]byte, error) {
	caKeyBlock, _ := pem.Decode(ca.Data["tls.key"])
	if caKeyBlock == nil {
		return nil, fmt.Errorf("failed decoding 'tls.key' data from secret %s", ca.Name)
	}
	priv, err := x509.ParseECPrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	caCertBlock, _ := pem.Decode(ca.Data["tls.crt"])
	if caCertBlock == nil {
		return nil, fmt.Errorf("failed decoding 'tls.crt' data from secret %s", ca.Name)
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, err
	}

	var usages []string
	for _, usage := range csr.Spec.Usages {
		usages = append(usages, string(usage))
	}

	certExpiryDuration := time.Second * time.Duration(*csr.Spec.ExpirationSeconds)
	durationUntilExpiry := caCert.NotAfter.Sub(time.Now()) //nolint:gosimple
	if durationUntilExpiry <= 0 {
		return nil, fmt.Errorf("the signer has expired: %v", caCert.NotAfter)
	}
	if durationUntilExpiry < certExpiryDuration {
		certExpiryDuration = durationUntilExpiry
	}

	policy := &config.Signing{
		Default: &config.SigningProfile{
			Usage:        usages,
			Expiry:       certExpiryDuration,
			ExpiryString: certExpiryDuration.String(),
		},
	}
	cfs, err := local.NewSigner(priv, caCert, x509.ECDSAWithSHA256, policy)
	if err != nil {
		return nil, err
	}

	certBytes, err := cfs.Sign(signer.SignRequest{Request: string(csr.Spec.Request)})
	if err != nil {
		return nil, err
	}
	return certBytes, nil
}

// EnsureCertificate creates a namespace/name Secret for subject signed by the CA in the
// mtlsCASecretNamespace/mtlsCASecretName Secret, or does nothing if a namespace/name Secret is
// already present. It returns a boolean indicating if it created a Secret and an error indicating
// any failures it encountered.
func EnsureCertificate[
	T interface {
		*operatorv1beta1.ControlPlane | *operatorv1beta1.DataPlane
		client.Object
	},
](
	ctx context.Context,
	owner T,
	subject string,
	mtlsCASecretNN types.NamespacedName,
	usages []certificatesv1.KeyUsage,
	cl client.Client,
	additionalMatchingLabels client.MatchingLabels,
) (op.CreatedUpdatedOrNoop, *corev1.Secret, error) {
	setCALogger(ctrlruntimelog.Log)

	// TODO: https://github.com/Kong/gateway-operator/pull/1101.
	// Use only new labels after several minor version of soak time.

	// Below we list both the Secrets with the new labels and the legacy labels
	// in order to support upgrades from older versions of the operator and perform
	// the reduction of the Secrets using the older labels.

	// Get the Secrets for the DataPlane using new labels.
	matchingLabels := k8sresources.GetManagedLabelForOwner(owner)
	for k, v := range additionalMatchingLabels {
		matchingLabels[k] = v
	}

	secrets, err := k8sutils.ListSecretsForOwner(ctx, cl, owner.GetUID(), matchingLabels)
	if err != nil {
		return op.Noop, nil, fmt.Errorf("failed listing Secrets for %T %s/%s: %w", owner, owner.GetNamespace(), owner.GetName(), err)
	}

	// Get the Secrets for the DataPlane using legacy labels.
	reqLegacyLabels, err := k8sresources.GetManagedLabelRequirementsForOwnerLegacy(owner)
	if err != nil {
		return op.Noop, nil, err
	}
	secretsLegacy, err := k8sutils.ListSecretsForOwner(
		ctx,
		cl,
		owner.GetUID(),
		&client.ListOptions{
			LabelSelector: labels.NewSelector().Add(reqLegacyLabels...),
		},
	)
	if err != nil {
		return op.Noop, nil, fmt.Errorf("failed listing Secrets for %T %s/%s: %w", owner, owner.GetNamespace(), owner.GetName(), err)
	}
	secrets = append(secrets, secretsLegacy...)

	count := len(secrets)
	if count > 1 {
		if err := k8sreduce.ReduceSecrets(ctx, cl, secrets, getPreDeleteHooks(owner)...); err != nil {
			return op.Noop, nil, err
		}
		return op.Noop, nil, errors.New("number of secrets reduced")
	}

	secretOpts := append(getSecretOpts(owner), matchingLabelsToSecretOpt(matchingLabels))
	generatedSecret := k8sresources.GenerateNewTLSSecret(owner, secretOpts...)

	// If there are no secrets yet, then create one.
	if count == 0 {
		return generateTLSDataSecret(ctx, generatedSecret, owner, subject, mtlsCASecretNN, usages, cl)
	}

	// Otherwise there is already 1 certificate matching specified selectors.
	existingSecret := &secrets[0]

	block, _ := pem.Decode(existingSecret.Data["tls.crt"])
	if block == nil {
		// The existing secret has a broken certificate, delete it and recreate it.
		if err := cl.Delete(ctx, existingSecret); err != nil {
			return op.Noop, nil, err
		}

		return generateTLSDataSecret(ctx, generatedSecret, owner, subject, mtlsCASecretNN, usages, cl)
	}

	// Check if existing certificate is for a different subject.
	// If that's the case, delete the old certificate and create a new one.
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return op.Noop, nil, err
	}
	if cert.Subject.CommonName != subject {
		if err := cl.Delete(ctx, existingSecret); err != nil {
			return op.Noop, nil, err
		}

		return generateTLSDataSecret(ctx, generatedSecret, owner, subject, mtlsCASecretNN, usages, cl)
	}

	var updated bool
	updated, existingSecret.ObjectMeta = k8sutils.EnsureObjectMetaIsUpdated(existingSecret.ObjectMeta, generatedSecret.ObjectMeta)
	if updated {
		if err := cl.Update(ctx, existingSecret); err != nil {
			return op.Noop, existingSecret, fmt.Errorf("failed updating secret %s: %w", existingSecret.Name, err)
		}
		return op.Updated, existingSecret, nil
	}
	return op.Noop, existingSecret, nil
}

func matchingLabelsToSecretOpt(ml client.MatchingLabels) k8sresources.SecretOpt {
	return func(a *corev1.Secret) {
		if a.Labels == nil {
			a.Labels = make(map[string]string)
		}
		for k, v := range ml {
			a.Labels[k] = v
		}
	}
}

// getPreDeleteHooks returns a list of pre-delete hooks for the given object type.
func getPreDeleteHooks[T interface {
	*operatorv1beta1.ControlPlane | *operatorv1beta1.DataPlane
	client.Object
},
](obj T,
) []k8sreduce.PreDeleteHook {
	switch any(obj).(type) {
	case *operatorv1beta1.DataPlane:
		return []k8sreduce.PreDeleteHook{dataplane.OwnedObjectPreDeleteHook}
	default:
		return nil
	}
}

// getSecretOpts returns a list of SecretOpt for the given object type.
func getSecretOpts[T interface {
	*operatorv1beta1.ControlPlane | *operatorv1beta1.DataPlane
	client.Object
},
](obj T,
) []k8sresources.SecretOpt {
	switch any(obj).(type) {
	case *operatorv1beta1.DataPlane:
		withDataPlaneOwnedFinalizer := func(s *corev1.Secret) {
			controllerutil.AddFinalizer(s, consts.DataPlaneOwnedWaitForOwnerFinalizer)
		}
		return []k8sresources.SecretOpt{withDataPlaneOwnedFinalizer}
	default:
		return nil
	}
}

// generateTLSDataSecret generates a TLS certificate data, fills the provided secret with
// that data and creates it using the k8s client.
// It returns a boolean indicating whether the secret has been created, the secret
// itself and an error.
func generateTLSDataSecret(
	ctx context.Context,
	generatedSecret *corev1.Secret,
	owner client.Object,
	subject string,
	mtlsCASecret types.NamespacedName,
	usages []certificatesv1.KeyUsage,
	k8sClient client.Client,
) (op.CreatedUpdatedOrNoop, *corev1.Secret, error) {
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   subject,
			Organization: []string{"Kong, Inc."},
			Country:      []string{"US"},
		},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		DNSNames:           []string{subject},
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return op.Noop, nil, err
	}

	der, err := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	if err != nil {
		return op.Noop, nil, err
	}

	// This is effectively a placeholder so long as we handle signing internally. When actually creating CSR resources,
	// this string is used by signers to filter which resources they pay attention to
	signerName := "gateway-operator.konghq.com/mtls"
	// TODO This creates certificates that last for 10 years as an arbitrarily long period for the alpha. A production-
	// ready implementation should use a shorter lifetime and rotate certificates. Rotation requires some mechanism to
	// recognize that certificates have expired (ideally without permissions to read Secrets across the cluster) and
	// to get Deployments to acknowledge them. For Kong, this requires a restart, as there's no way to force a reload
	// of updated files on disk.
	expiration := int32(315400000)

	csr := certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: owner.GetNamespace(),
			Name:      owner.GetName(),
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request: pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE REQUEST",
				Bytes: der,
			}),
			SignerName:        signerName,
			ExpirationSeconds: &expiration,
			Usages:            usages,
		},
	}

	ca := &corev1.Secret{}
	err = k8sClient.Get(ctx, mtlsCASecret, ca)
	if err != nil {
		return op.Noop, nil, err
	}

	signed, err := signCertificate(csr, ca)
	if err != nil {
		return op.Noop, nil, err
	}
	privDer, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return op.Noop, nil, err
	}

	generatedSecret.Data = map[string][]byte{
		"ca.crt":  ca.Data["tls.crt"],
		"tls.crt": signed,
		"tls.key": pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: privDer,
		}),
	}

	err = k8sClient.Create(ctx, generatedSecret)
	if err != nil {
		return op.Noop, nil, err
	}

	return op.Created, generatedSecret, nil
}

// GetManagedLabelForServiceSecret returns a label selector for the ServiceSecret.
func GetManagedLabelForServiceSecret(svcNN types.NamespacedName) client.MatchingLabels {
	return client.MatchingLabels{
		consts.ServiceSecretLabel: svcNN.Name,
	}
}

// -----------------------------------------------------------------------------
// Private Functions - Container Images
// -----------------------------------------------------------------------------

// ensureContainerImageUpdated ensures that the provided container is
// configured with a container image consistent with the image and
// image version provided. The image and version can be provided as
// nil when not wanted.
func ensureContainerImageUpdated(container *corev1.Container, imageVersionStr string) (updated bool, err error) {
	// Can't update a with an empty image.
	if imageVersionStr == "" {
		return false, fmt.Errorf("can't update container image with an empty image")
	}

	imageParts := strings.Split(container.Image, ":")
	if len(imageParts) > 3 {
		err = fmt.Errorf("invalid container image found: %s", container.Image)
		return
	}

	// This is a special case for registries that specify a non default port,
	// e.g. localhost:5000 or myregistry.io:8000. We do look for '/' since the
	// container.Image will contain it as a separator between the registry+image
	// and the version.
	if len(imageParts) == 3 {
		if !strings.Contains(container.Image, "/") {
			return false, fmt.Errorf("invalid container image found: %s", container.Image)
		}

		containerImageURL := imageParts[0] + imageParts[1]
		u, err := url.Parse(containerImageURL)
		if err != nil {
			return false, fmt.Errorf("invalid registry URL %s: %w", containerImageURL, err)
		}
		containerImageURL = u.String()
		container.Image = containerImageURL + ":" + imageParts[2]
		updated = true
	}

	if imageVersionStr != container.Image {
		container.Image = imageVersionStr
		updated = true
	}

	return
}
