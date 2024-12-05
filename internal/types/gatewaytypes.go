package types

import (
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type (
	Gateway              = gatewayv1.Gateway
	GatewayAddress       = gatewayv1.GatewayAddress
	GatewaySpec          = gatewayv1.GatewaySpec
	GatewayStatusAddress = gatewayv1.GatewayStatusAddress
	Listener             = gatewayv1.Listener
	HTTPRoute            = gatewayv1.HTTPRoute
	HTTPRouteSpec        = gatewayv1.HTTPRouteSpec
	HTTPRouteList        = gatewayv1.HTTPRouteList
	ParentReference      = gatewayv1.ParentReference
	CommonRouteSpec      = gatewayv1.CommonRouteSpec
	Kind                 = gatewayv1.Kind
	Namespace            = gatewayv1.Namespace
	Group                = gatewayv1.Group
	AllowedRoutes        = gatewayv1.AllowedRoutes
	RouteGroupKind       = gatewayv1.RouteGroupKind
	RouteNamespaces      = gatewayv1.RouteNamespaces
	ObjectName           = gatewayv1.ObjectName
	SectionName          = gatewayv1.SectionName
	PortNumber           = gatewayv1.PortNumber
)

var GroupVersion = gatewayv1.GroupVersion

const (
	HTTPProtocolType = gatewayv1.HTTPProtocolType

	NamespacesFromAll      = gatewayv1.NamespacesFromAll
	NamespacesFromSame     = gatewayv1.NamespacesFromSame
	NamespacesFromSelector = gatewayv1.NamespacesFromSelector
)
