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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	apisv1beta1 "github.com/kong/gateway-operator/pkg/clientset/typed/apis/v1beta1"
	gentype "k8s.io/client-go/gentype"
)

// fakeGatewayConfigurations implements GatewayConfigurationInterface
type fakeGatewayConfigurations struct {
	*gentype.FakeClientWithList[*v1beta1.GatewayConfiguration, *v1beta1.GatewayConfigurationList]
	Fake *FakeApisV1beta1
}

func newFakeGatewayConfigurations(fake *FakeApisV1beta1, namespace string) apisv1beta1.GatewayConfigurationInterface {
	return &fakeGatewayConfigurations{
		gentype.NewFakeClientWithList[*v1beta1.GatewayConfiguration, *v1beta1.GatewayConfigurationList](
			fake.Fake,
			namespace,
			v1beta1.SchemeGroupVersion.WithResource("gatewayconfigurations"),
			v1beta1.SchemeGroupVersion.WithKind("GatewayConfiguration"),
			func() *v1beta1.GatewayConfiguration { return &v1beta1.GatewayConfiguration{} },
			func() *v1beta1.GatewayConfigurationList { return &v1beta1.GatewayConfigurationList{} },
			func(dst, src *v1beta1.GatewayConfigurationList) { dst.ListMeta = src.ListMeta },
			func(list *v1beta1.GatewayConfigurationList) []*v1beta1.GatewayConfiguration {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1beta1.GatewayConfigurationList, items []*v1beta1.GatewayConfiguration) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
