/*
Copyright 2018 The Kubernetes Authors.

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

package controller

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kong/kubernetes-ingress-controller/internal/ingress"
)

func TestGetEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		svc    *corev1.Service
		port   *corev1.ServicePort
		proto  corev1.Protocol
		fn     func(*corev1.Service) (*corev1.Endpoints, error)
		result []ingress.Endpoint
	}{
		{
			"no service should return 0 endpoints",
			nil,
			nil,
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				return nil, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"no service port should return 0 endpoints",
			&corev1.Service{},
			nil,
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				return nil, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"a service without endpoints should return 0 endpoints",
			&corev1.Service{},
			&corev1.ServicePort{Name: "default"},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				return &corev1.Endpoints{}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"a service type ServiceTypeExternalName service with an invalid port should return 0 endpoints",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
				},
			},
			&corev1.ServicePort{Name: "default"},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				return &corev1.Endpoints{}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"a service type ServiceTypeExternalName with a valid port should return one endpoint",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:         corev1.ServiceTypeExternalName,
					ExternalName: "10.0.0.1.xip.io",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				return &corev1.Endpoints{}, nil
			},
			[]ingress.Endpoint{
				{
					Address:     "10.0.0.1.xip.io",
					Port:        "80",
					MaxFails:    0,
					FailTimeout: 0,
				},
			},
		},
		{
			"should return no endpoints when there is an error searching for endpoints",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				return nil, fmt.Errorf("unexpected error")
			},
			[]ingress.Endpoint{},
		},
		{
			"should return no endpoints when the protocol does not match",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				nodeName := "dummy"
				return &corev1.Endpoints{
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP:       "1.1.1.1",
									NodeName: &nodeName,
								},
							},
							Ports: []corev1.EndpointPort{
								{
									Protocol: corev1.ProtocolUDP,
								},
							},
						},
					},
				}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"should return no endpoints when there is no ready Addresses",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				nodeName := "dummy"
				return &corev1.Endpoints{
					Subsets: []corev1.EndpointSubset{
						{
							NotReadyAddresses: []corev1.EndpointAddress{
								{
									IP:       "1.1.1.1",
									NodeName: &nodeName,
								},
							},
							Ports: []corev1.EndpointPort{
								{
									Protocol: corev1.ProtocolUDP,
								},
							},
						},
					},
				}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"should return no endpoints when the name of the port name do not match any port in the endpoint Subsets",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				nodeName := "dummy"
				return &corev1.Endpoints{
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP:       "1.1.1.1",
									NodeName: &nodeName,
								},
							},
							Ports: []corev1.EndpointPort{
								{
									Protocol: corev1.ProtocolTCP,
									Port:     int32(80),
									Name:     "another-name",
								},
							},
						},
					},
				}, nil
			},
			[]ingress.Endpoint{},
		},
		{
			"should return one endpoint when the name of the port name match a port in the endpoint Subsets",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromInt(80),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "default",
				TargetPort: intstr.FromInt(80),
			},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				nodeName := "dummy"
				return &corev1.Endpoints{
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP:       "1.1.1.1",
									NodeName: &nodeName,
								},
							},
							Ports: []corev1.EndpointPort{
								{
									Protocol: corev1.ProtocolTCP,
									Port:     int32(80),
									Name:     "default",
								},
							},
						},
					},
				}, nil
			},
			[]ingress.Endpoint{
				{
					Address:     "1.1.1.1",
					Port:        "80",
					MaxFails:    0,
					FailTimeout: 0,
				},
			},
		},
		{
			"should return one endpoint when the name of the port name match more than one port in the endpoint Subsets",
			&corev1.Service{
				Spec: corev1.ServiceSpec{
					Type:      corev1.ServiceTypeClusterIP,
					ClusterIP: "1.1.1.1",
					Ports: []corev1.ServicePort{
						{
							Name:       "default",
							TargetPort: intstr.FromString("port-1"),
						},
					},
				},
			},
			&corev1.ServicePort{
				Name:       "port-1",
				TargetPort: intstr.FromString("port-1"),
			},
			corev1.ProtocolTCP,
			func(*corev1.Service) (*corev1.Endpoints, error) {
				nodeName := "dummy"
				return &corev1.Endpoints{
					Subsets: []corev1.EndpointSubset{
						{
							Addresses: []corev1.EndpointAddress{
								{
									IP:       "1.1.1.1",
									NodeName: &nodeName,
								},
							},
							Ports: []corev1.EndpointPort{
								{
									Name:     "port-1",
									Protocol: corev1.ProtocolTCP,
									Port:     80,
								},
								{
									Name:     "port-1",
									Protocol: corev1.ProtocolTCP,
									Port:     80,
								},
							},
						},
					},
				}, nil
			},
			[]ingress.Endpoint{
				{
					Address:     "1.1.1.1",
					Port:        "80",
					MaxFails:    0,
					FailTimeout: 0,
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := getEndpoints(testCase.svc, testCase.port, testCase.proto, testCase.fn)
			if len(testCase.result) != len(result) {
				t.Errorf("expected %v Endpoints but got %v", testCase.result, len(result))
			}
		})
	}
}
