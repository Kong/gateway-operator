package parser

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kong/go-kong/kong"
	"github.com/kong/kubernetes-ingress-controller/internal/ingress/annotations"
	"github.com/kong/kubernetes-ingress-controller/internal/ingress/controller/parser/kongstate"
	"github.com/kong/kubernetes-ingress-controller/internal/ingress/store"
	"github.com/kong/kubernetes-ingress-controller/internal/ingress/utils"
	configurationv1beta1 "github.com/kong/kubernetes-ingress-controller/pkg/apis/configuration/v1beta1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knative "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

func parseAll(log logrus.FieldLogger, s store.Storer) ingressRules {
	ings := s.ListIngresses()
	tcpIngresses, err := s.ListTCPIngresses()
	if err != nil {
		log.Errorf("failed to list TCPIngresses: %v", err)
	}
	parsedIngress := parseIngressRules(log, ings, tcpIngresses)

	knativeIngresses, err := s.ListKnativeIngresses()
	if err != nil {
		log.Errorf("failed to list Knative Ingresses: %v", err)
	}
	parsedKnative := parseKnativeIngressRules(knativeIngresses)

	return mergeIngressRules(&parsedIngress, &parsedKnative)
}

// Build creates a Kong configuration from Ingress and Custom resources
// defined in Kuberentes.
// It throws an error if there is an error returned from client-go.
func Build(log logrus.FieldLogger, s store.Storer) (*kongstate.KongState, error) {
	parsedAll := parseAll(log, s)
	parsedAll.populateServices(log, s)

	var result kongstate.KongState
	// add the routes and services to the state
	for _, service := range parsedAll.ServiceNameToServices {
		result.Services = append(result.Services, service)
	}

	// generate Upstreams and Targets from service defs
	result.Upstreams = getUpstreams(log, s, parsedAll.ServiceNameToServices)

	// merge KongIngress with Routes, Services and Upstream
	result.FillOverrides(log, s)

	// generate consumers and credentials
	result.FillConsumersAndCredentials(log, s)

	// process annotation plugins
	result.FillPlugins(log, s)

	// generate Certificates and SNIs
	result.Certificates = getCerts(log, s, parsedAll.SecretNameToSNIs)

	// populate CA certificates in Kong
	var err error
	caCertSecrets, err := s.ListCACerts()
	if err != nil {
		return nil, err
	}
	result.CACertificates = toCACerts(log, caCertSecrets)

	return &result, nil
}

func toCACerts(log logrus.FieldLogger, caCertSecrets []*corev1.Secret) []kong.CACertificate {
	var caCerts []kong.CACertificate
	for _, certSecret := range caCertSecrets {
		secretName := certSecret.Namespace + "/" + certSecret.Name

		idbytes, idExists := certSecret.Data["id"]
		log = log.WithFields(logrus.Fields{
			"secret_name":      secretName,
			"secret_namespace": certSecret.Namespace,
		})
		if !idExists {
			log.Errorf("invalid CA certificate: missing 'id' field in data")
			continue
		}

		caCertbytes, certExists := certSecret.Data["cert"]
		if !certExists {
			log.Errorf("invalid CA certificate: missing 'cert' field in data")
			continue
		}

		pemBlock, _ := pem.Decode(caCertbytes)
		if pemBlock == nil {
			log.Errorf("invalid CA certificate: invalid PEM block")
			continue
		}
		x509Cert, err := x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			log.Errorf("invalid CA certificate: failed to parse certificate: %v", err)
			continue
		}
		if !x509Cert.IsCA {
			log.Errorf("invalid CA certificate: certificate is missing the 'CA' basic constraint: %v", err)
			continue
		}

		caCerts = append(caCerts, kong.CACertificate{
			ID:   kong.String(string(idbytes)),
			Cert: kong.String(string(caCertbytes)),
		})
	}

	return caCerts
}

func knativeIngressToNetworkingTLS(tls []knative.IngressTLS) []networking.IngressTLS {
	var result []networking.IngressTLS

	for _, t := range tls {
		result = append(result, networking.IngressTLS{
			Hosts:      t.Hosts,
			SecretName: t.SecretName,
		})
	}
	return result
}

func tcpIngressToNetworkingTLS(tls []configurationv1beta1.IngressTLS) []networking.IngressTLS {
	var result []networking.IngressTLS

	for _, t := range tls {
		result = append(result, networking.IngressTLS{
			Hosts:      t.Hosts,
			SecretName: t.SecretName,
		})
	}
	return result
}

func parseKnativeIngressRules(
	ingressList []*knative.Ingress) ingressRules {

	sort.SliceStable(ingressList, func(i, j int) bool {
		return ingressList[i].CreationTimestamp.Before(
			&ingressList[j].CreationTimestamp)
	})

	services := map[string]kongstate.Service{}
	secretToSNIs := newSecretNameToSNIs()

	for i := 0; i < len(ingressList); i++ {
		ingress := *ingressList[i]
		ingressSpec := ingress.Spec

		secretToSNIs.addFromIngressTLS(knativeIngressToNetworkingTLS(ingress.Spec.TLS), ingress.Namespace)
		for i, rule := range ingressSpec.Rules {
			hosts := rule.Hosts
			if rule.HTTP == nil {
				continue
			}
			for j, rule := range rule.HTTP.Paths {
				path := rule.Path

				if path == "" {
					path = "/"
				}
				r := kongstate.Route{
					Route: kong.Route{
						// TODO Figure out a way to name the routes
						// This is not a stable scheme
						// 1. If a user adds a route in the middle,
						// due to a shift, all the following routes will
						// be PATCHED
						// 2. Is it guaranteed that the order is stable?
						// Meaning, the routes will always appear in the same
						// order?
						Name:          kong.String(ingress.Namespace + "." + ingress.Name + "." + strconv.Itoa(i) + strconv.Itoa(j)),
						Paths:         kong.StringSlice(path),
						StripPath:     kong.Bool(false),
						PreserveHost:  kong.Bool(true),
						Protocols:     kong.StringSlice("http", "https"),
						RegexPriority: kong.Int(0),
					},
				}
				r.Hosts = kong.StringSlice(hosts...)

				knativeBackend := knativeSelectSplit(rule.Splits)
				serviceName := knativeBackend.ServiceNamespace + "." +
					knativeBackend.ServiceName + "." +
					knativeBackend.ServicePort.String()
				serviceHost := knativeBackend.ServiceName + "." +
					knativeBackend.ServiceNamespace + "." +
					knativeBackend.ServicePort.String() + ".svc"
				service, ok := services[serviceName]
				if !ok {

					var headers []string
					for key, value := range knativeBackend.AppendHeaders {
						headers = append(headers, key+":"+value)
					}
					for key, value := range rule.AppendHeaders {
						headers = append(headers, key+":"+value)
					}

					service = kongstate.Service{
						Service: kong.Service{
							Name:           kong.String(serviceName),
							Host:           kong.String(serviceHost),
							Port:           kong.Int(80),
							Protocol:       kong.String("http"),
							Path:           kong.String("/"),
							ConnectTimeout: kong.Int(60000),
							ReadTimeout:    kong.Int(60000),
							WriteTimeout:   kong.Int(60000),
							Retries:        kong.Int(5),
						},
						Namespace: ingress.Namespace,
						Backend: kongstate.ServiceBackend{
							Name: knativeBackend.ServiceName,
							Port: knativeBackend.ServicePort,
						},
					}
					if len(headers) > 0 {
						service.Plugins = append(service.Plugins, kong.Plugin{
							Name: kong.String("request-transformer"),
							Config: kong.Configuration{
								"add": map[string]interface{}{
									"headers": headers,
								},
							},
						})
					}
				}
				service.Routes = append(service.Routes, r)
				services[serviceName] = service
			}
		}
	}

	return ingressRules{
		ServiceNameToServices: services,
		SecretNameToSNIs:      secretToSNIs,
	}
}

func knativeSelectSplit(splits []knative.IngressBackendSplit) knative.IngressBackendSplit {
	if len(splits) == 0 {
		return knative.IngressBackendSplit{}
	}
	res := splits[0]
	maxPercentage := splits[0].Percent
	if len(splits) == 1 {
		return res
	}
	for i := 1; i < len(splits); i++ {
		if splits[i].Percent > maxPercentage {
			res = splits[i]
			maxPercentage = res.Percent
		}
	}
	return res
}

func parseIngressRules(
	log logrus.FieldLogger,
	ingressList []*networking.Ingress,
	tcpIngressList []*configurationv1beta1.TCPIngress) ingressRules {

	sort.SliceStable(ingressList, func(i, j int) bool {
		return ingressList[i].CreationTimestamp.Before(
			&ingressList[j].CreationTimestamp)
	})

	sort.SliceStable(tcpIngressList, func(i, j int) bool {
		return tcpIngressList[i].CreationTimestamp.Before(
			&tcpIngressList[j].CreationTimestamp)
	})

	// generate the following:
	// Services and Routes
	var allDefaultBackends []networking.Ingress
	secretNameToSNIs := newSecretNameToSNIs()
	serviceNameToServices := make(map[string]kongstate.Service)

	for i := 0; i < len(ingressList); i++ {
		ingress := *ingressList[i]
		ingressSpec := ingress.Spec
		log = log.WithFields(logrus.Fields{
			"ingress_namespace": ingress.Namespace,
			"ingress_name":      ingress.Name,
		})

		if ingressSpec.Backend != nil {
			allDefaultBackends = append(allDefaultBackends, ingress)

		}

		secretNameToSNIs.addFromIngressTLS(ingressSpec.TLS, ingress.Namespace)

		for i, rule := range ingressSpec.Rules {
			host := rule.Host
			if rule.HTTP == nil {
				continue
			}
			for j, rule := range rule.HTTP.Paths {
				path := rule.Path

				if strings.Contains(path, "//") {
					log.Errorf("rule skipped: invalid path: '%v'", path)
					continue
				}
				if path == "" {
					path = "/"
				}
				r := kongstate.Route{
					Ingress: ingress,
					Route: kong.Route{
						// TODO Figure out a way to name the routes
						// This is not a stable scheme
						// 1. If a user adds a route in the middle,
						// due to a shift, all the following routes will
						// be PATCHED
						// 2. Is it guaranteed that the order is stable?
						// Meaning, the routes will always appear in the same
						// order?
						Name:          kong.String(ingress.Namespace + "." + ingress.Name + "." + strconv.Itoa(i) + strconv.Itoa(j)),
						Paths:         kong.StringSlice(path),
						StripPath:     kong.Bool(false),
						PreserveHost:  kong.Bool(true),
						Protocols:     kong.StringSlice("http", "https"),
						RegexPriority: kong.Int(0),
					},
				}
				if host != "" {
					r.Hosts = kong.StringSlice(host)
				}

				serviceName := ingress.Namespace + "." +
					rule.Backend.ServiceName + "." +
					rule.Backend.ServicePort.String()
				service, ok := serviceNameToServices[serviceName]
				if !ok {
					service = kongstate.Service{
						Service: kong.Service{
							Name: kong.String(serviceName),
							Host: kong.String(rule.Backend.ServiceName +
								"." + ingress.Namespace + "." +
								rule.Backend.ServicePort.String() + ".svc"),
							Port:           kong.Int(80),
							Protocol:       kong.String("http"),
							Path:           kong.String("/"),
							ConnectTimeout: kong.Int(60000),
							ReadTimeout:    kong.Int(60000),
							WriteTimeout:   kong.Int(60000),
							Retries:        kong.Int(5),
						},
						Namespace: ingress.Namespace,
						Backend: kongstate.ServiceBackend{
							Name: rule.Backend.ServiceName,
							Port: rule.Backend.ServicePort,
						},
					}
				}
				service.Routes = append(service.Routes, r)
				serviceNameToServices[serviceName] = service
			}
		}
	}

	for i := 0; i < len(tcpIngressList); i++ {
		ingress := *tcpIngressList[i]
		ingressSpec := ingress.Spec

		log = log.WithFields(logrus.Fields{
			"tcpingress_namespace": ingress.Namespace,
			"tcpingress_name":      ingress.Name,
		})

		secretNameToSNIs.addFromIngressTLS(tcpIngressToNetworkingTLS(ingressSpec.TLS), ingress.Namespace)

		for i, rule := range ingressSpec.Rules {

			if rule.Port <= 0 {
				log.Errorf("invalid TCPIngress: invalid port: %v", rule.Port)
				continue
			}
			r := kongstate.Route{
				IsTCP:      true,
				TCPIngress: ingress,
				Route: kong.Route{
					// TODO Figure out a way to name the routes
					// This is not a stable scheme
					// 1. If a user adds a route in the middle,
					// due to a shift, all the following routes will
					// be PATCHED
					// 2. Is it guaranteed that the order is stable?
					// Meaning, the routes will always appear in the same
					// order?
					Name:      kong.String(ingress.Namespace + "." + ingress.Name + "." + strconv.Itoa(i)),
					Protocols: kong.StringSlice("tcp", "tls"),
					Destinations: []*kong.CIDRPort{
						{
							Port: kong.Int(rule.Port),
						},
					},
				},
			}
			host := rule.Host
			if host != "" {
				r.SNIs = kong.StringSlice(host)
			}
			if rule.Backend.ServiceName == "" {
				log.Errorf("invalid TCPIngress: empty serviceName")
				continue
			}
			if rule.Backend.ServicePort <= 0 {
				log.Errorf("invalid TCPIngress: invalid servicePort: %v", rule.Backend.ServicePort)
				continue
			}

			serviceName := ingress.Namespace + "." +
				rule.Backend.ServiceName + "." +
				strconv.Itoa(rule.Backend.ServicePort)
			service, ok := serviceNameToServices[serviceName]
			if !ok {
				service = kongstate.Service{
					Service: kong.Service{
						Name: kong.String(serviceName),
						Host: kong.String(rule.Backend.ServiceName +
							"." + ingress.Namespace + "." +
							strconv.Itoa(rule.Backend.ServicePort) + ".svc"),
						Port:           kong.Int(80),
						Protocol:       kong.String("tcp"),
						ConnectTimeout: kong.Int(60000),
						ReadTimeout:    kong.Int(60000),
						WriteTimeout:   kong.Int(60000),
						Retries:        kong.Int(5),
					},
					Namespace: ingress.Namespace,
					Backend: kongstate.ServiceBackend{
						Name: rule.Backend.ServiceName,
						Port: intstr.FromInt(rule.Backend.ServicePort),
					},
				}
			}
			service.Routes = append(service.Routes, r)
			serviceNameToServices[serviceName] = service
		}
	}

	sort.SliceStable(allDefaultBackends, func(i, j int) bool {
		return allDefaultBackends[i].CreationTimestamp.Before(&allDefaultBackends[j].CreationTimestamp)
	})

	// Process the default backend
	if len(allDefaultBackends) > 0 {
		ingress := allDefaultBackends[0]
		defaultBackend := allDefaultBackends[0].Spec.Backend
		serviceName := allDefaultBackends[0].Namespace + "." +
			defaultBackend.ServiceName + "." +
			defaultBackend.ServicePort.String()
		service, ok := serviceNameToServices[serviceName]
		if !ok {
			service = kongstate.Service{
				Service: kong.Service{
					Name: kong.String(serviceName),
					Host: kong.String(defaultBackend.ServiceName + "." +
						ingress.Namespace + "." +
						defaultBackend.ServicePort.String() + ".svc"),
					Port:           kong.Int(80),
					Protocol:       kong.String("http"),
					ConnectTimeout: kong.Int(60000),
					ReadTimeout:    kong.Int(60000),
					WriteTimeout:   kong.Int(60000),
					Retries:        kong.Int(5),
				},
				Namespace: ingress.Namespace,
				Backend: kongstate.ServiceBackend{
					Name: defaultBackend.ServiceName,
					Port: defaultBackend.ServicePort,
				},
			}
		}
		r := kongstate.Route{
			Ingress: ingress,
			Route: kong.Route{
				Name:          kong.String(ingress.Namespace + "." + ingress.Name),
				Paths:         kong.StringSlice("/"),
				StripPath:     kong.Bool(false),
				PreserveHost:  kong.Bool(true),
				Protocols:     kong.StringSlice("http", "https"),
				RegexPriority: kong.Int(0),
			},
		}
		service.Routes = append(service.Routes, r)
		serviceNameToServices[serviceName] = service
	}

	return ingressRules{
		SecretNameToSNIs:      secretNameToSNIs,
		ServiceNameToServices: serviceNameToServices,
	}
}

func getUpstreams(
	log logrus.FieldLogger, s store.Storer, serviceMap map[string]kongstate.Service) []kongstate.Upstream {
	var upstreams []kongstate.Upstream
	for _, service := range serviceMap {
		upstreamName := service.Backend.Name + "." + service.Namespace + "." + service.Backend.Port.String() + ".svc"
		upstream := kongstate.Upstream{
			Upstream: kong.Upstream{
				Name: kong.String(upstreamName),
			},
			Service: service,
		}
		targets := getServiceEndpoints(log, s, service.K8sService,
			service.Backend.Port.String())
		upstream.Targets = targets
		upstreams = append(upstreams, upstream)
	}
	return upstreams
}

func getCertFromSecret(secret *corev1.Secret) (string, string, error) {
	certData, okcert := secret.Data[corev1.TLSCertKey]
	keyData, okkey := secret.Data[corev1.TLSPrivateKeyKey]

	if !okcert || !okkey {
		return "", "", fmt.Errorf("no keypair could be found in"+
			" secret '%v/%v'", secret.Namespace, secret.Name)
	}

	cert := strings.TrimSpace(bytes.NewBuffer(certData).String())
	key := strings.TrimSpace(bytes.NewBuffer(keyData).String())

	_, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return "", "", fmt.Errorf("parsing TLS key-pair in secret '%v/%v': %v",
			secret.Namespace, secret.Name, err)
	}

	return cert, key, nil
}

func getCerts(log logrus.FieldLogger, s store.Storer, secretsToSNIs map[string][]string) []kongstate.Certificate {
	snisAdded := make(map[string]bool)
	// map of cert public key + private key to certificate
	type certWrapper struct {
		cert              kong.Certificate
		CreationTimestamp metav1.Time
	}
	certs := make(map[string]certWrapper)

	for secretKey, SNIs := range secretsToSNIs {
		namespaceName := strings.Split(secretKey, "/")
		secret, err := s.GetSecret(namespaceName[0], namespaceName[1])
		if err != nil {
			log.WithFields(logrus.Fields{
				"secret_name":      namespaceName[1],
				"secret_namespace": namespaceName[0],
			}).Logger.Errorf("failed to fetch secret: %v", err)
			continue
		}
		cert, key, err := getCertFromSecret(secret)
		if err != nil {
			log.WithFields(logrus.Fields{
				"secret_name":      namespaceName[1],
				"secret_namespace": namespaceName[0],
			}).Logger.Errorf("failed to construct certificate from secret: %v", err)
			continue
		}
		kongCert, ok := certs[cert+key]
		if !ok {
			kongCert = certWrapper{
				cert: kong.Certificate{
					ID:   kong.String(string(secret.UID)),
					Cert: kong.String(cert),
					Key:  kong.String(key),
				},
				CreationTimestamp: secret.CreationTimestamp,
			}
		} else {
			if kongCert.CreationTimestamp.After(secret.CreationTimestamp.Time) {
				kongCert.cert.ID = kong.String(string(secret.UID))
				kongCert.CreationTimestamp = secret.CreationTimestamp
			}
		}

		for _, sni := range SNIs {
			if !snisAdded[sni] {
				snisAdded[sni] = true
				kongCert.cert.SNIs = append(kongCert.cert.SNIs, kong.String(sni))
			}
		}
		certs[cert+key] = kongCert
	}
	var res []kongstate.Certificate
	for _, cert := range certs {
		res = append(res, kongstate.Certificate{Certificate: cert.cert})
	}
	return res
}

func getServiceEndpoints(log logrus.FieldLogger, s store.Storer, svc corev1.Service,
	backendPort string) []kongstate.Target {
	var targets []kongstate.Target
	var endpoints []utils.Endpoint
	var servicePort corev1.ServicePort

	log = log.WithFields(logrus.Fields{
		"service_name":      svc.Name,
		"service_namespace": svc.Namespace,
	})

	for _, port := range svc.Spec.Ports {
		// targetPort could be a string, use the name or the port (int)
		if strconv.Itoa(int(port.Port)) == backendPort ||
			port.TargetPort.String() == backendPort ||
			port.Name == backendPort {
			servicePort = port
			break
		}
	}

	// Ingress with an ExternalName service and no port defined in the service.
	if len(svc.Spec.Ports) == 0 &&
		svc.Spec.Type == corev1.ServiceTypeExternalName {
		// nolint: gosec
		externalPort, err := strconv.Atoi(backendPort)
		if err != nil {
			log.Warningf("invalid ExternalName Service (only numeric ports allowed): %v", backendPort)
			return targets
		}

		servicePort = corev1.ServicePort{
			Protocol:   "TCP",
			Port:       int32(externalPort),
			TargetPort: intstr.FromString(backendPort),
		}
	}

	endpoints = getEndpoints(log, &svc, &servicePort,
		corev1.ProtocolTCP, s.GetEndpointsForService)
	if len(endpoints) == 0 {
		log.Warningf("no active endpionts")
	}
	for _, endpoint := range endpoints {
		target := kongstate.Target{
			Target: kong.Target{
				Target: kong.String(endpoint.Address + ":" + endpoint.Port),
			},
		}
		targets = append(targets, target)
	}
	return targets
}

// getEndpoints returns a list of <endpoint ip>:<port> for a given service/target port combination.
func getEndpoints(
	log logrus.FieldLogger,
	s *corev1.Service,
	port *corev1.ServicePort,
	proto corev1.Protocol,
	getEndpoints func(string, string) (*corev1.Endpoints, error),
) []utils.Endpoint {

	upsServers := []utils.Endpoint{}

	if s == nil || port == nil {
		return upsServers
	}

	log = log.WithFields(logrus.Fields{
		"service_name":      s.Name,
		"service_namespace": s.Namespace,
		"service_port":      port.String(),
	})

	// avoid duplicated upstream servers when the service
	// contains multiple port definitions sharing the same
	// targetport.
	adus := make(map[string]bool)

	// ExternalName services
	if s.Spec.Type == corev1.ServiceTypeExternalName {
		log.Debug("found service of type=ExternalName")

		targetPort := port.TargetPort.IntValue()
		// check for invalid port value
		if targetPort <= 0 {
			log.Errorf("invalid service: invalid port: %v", targetPort)
			return upsServers
		}

		return append(upsServers, utils.Endpoint{
			Address: s.Spec.ExternalName,
			Port:    fmt.Sprintf("%v", targetPort),
		})
	}
	if annotations.HasServiceUpstreamAnnotation(s.Annotations) {
		return append(upsServers, utils.Endpoint{
			Address: s.Name + "." + s.Namespace + ".svc",
			Port:    fmt.Sprintf("%v", port.Port),
		})

	}

	log.Debugf("fetching endpoints")
	ep, err := getEndpoints(s.Namespace, s.Name)
	if err != nil {
		log.Errorf("failed to fetch endpoints: %v", err)
		return upsServers
	}

	for _, ss := range ep.Subsets {
		for _, epPort := range ss.Ports {

			if !reflect.DeepEqual(epPort.Protocol, proto) {
				continue
			}

			var targetPort int32

			if port.Name == "" {
				// port.Name is optional if there is only one port
				targetPort = epPort.Port
			} else if port.Name == epPort.Name {
				targetPort = epPort.Port
			}

			// check for invalid port value
			if targetPort <= 0 {
				continue
			}

			for _, epAddress := range ss.Addresses {
				ep := fmt.Sprintf("%v:%v", epAddress.IP, targetPort)
				if _, exists := adus[ep]; exists {
					continue
				}
				ups := utils.Endpoint{
					Address: epAddress.IP,
					Port:    fmt.Sprintf("%v", targetPort),
				}
				upsServers = append(upsServers, ups)
				adus[ep] = true
			}
		}
	}

	log.Debugf("found endpoints: %v", upsServers)
	return upsServers
}
