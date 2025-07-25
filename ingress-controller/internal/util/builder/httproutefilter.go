package builder

import (
	"github.com/samber/lo"

	"github.com/kong/kong-operator/ingress-controller/internal/gatewayapi"
)

// HTTPRouteFilterBuilder is a builder for gateway api HTTPRouteMatch.
// Primarily used for testing.
type HTTPRouteFilterBuilder struct {
	httpRouteFilter gatewayapi.HTTPRouteFilter
}

func (b *HTTPRouteFilterBuilder) Build() gatewayapi.HTTPRouteFilter {
	return b.httpRouteFilter
}

// NewHTTPRouteRequestRedirectFilter builds a request redirect HTTPRoute filter.
func NewHTTPRouteRequestRedirectFilter() *HTTPRouteFilterBuilder {
	filter := gatewayapi.HTTPRouteFilter{
		Type:            gatewayapi.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &gatewayapi.HTTPRequestRedirectFilter{},
	}
	return &HTTPRouteFilterBuilder{httpRouteFilter: filter}
}

// WithRequestRedirectScheme sets scheme of request redirect filter.
func (b *HTTPRouteFilterBuilder) WithRequestRedirectScheme(scheme string) *HTTPRouteFilterBuilder {
	if b.httpRouteFilter.Type != gatewayapi.HTTPRouteFilterRequestRedirect ||
		b.httpRouteFilter.RequestRedirect == nil {
		return b
	}

	b.httpRouteFilter.RequestRedirect.Scheme = lo.ToPtr(scheme)
	return b
}

// WithRequestRedirectHost sets host of request redirect filter.
func (b *HTTPRouteFilterBuilder) WithRequestRedirectHost(host string) *HTTPRouteFilterBuilder {
	if b.httpRouteFilter.Type != gatewayapi.HTTPRouteFilterRequestRedirect ||
		b.httpRouteFilter.RequestRedirect == nil {
		return b
	}

	preciseHost := (gatewayapi.PreciseHostname)(host)
	b.httpRouteFilter.RequestRedirect.Hostname = lo.ToPtr(preciseHost)
	return b
}

// WithRequestRedirectStatusCode sets status code of response in request redirect filter.
func (b *HTTPRouteFilterBuilder) WithRequestRedirectStatusCode(code int) *HTTPRouteFilterBuilder {
	if b.httpRouteFilter.Type != gatewayapi.HTTPRouteFilterRequestRedirect ||
		b.httpRouteFilter.RequestRedirect == nil {
		return b
	}

	b.httpRouteFilter.RequestRedirect.StatusCode = lo.ToPtr(code)
	return b
}

func (b *HTTPRouteFilterBuilder) WithRequestRedirectPathModifier(modifierType gatewayapi.HTTPPathModifierType, path string) *HTTPRouteFilterBuilder {
	if b.httpRouteFilter.Type != gatewayapi.HTTPRouteFilterRequestRedirect ||
		b.httpRouteFilter.RequestRedirect == nil {
		return b
	}
	if modifierType == gatewayapi.FullPathHTTPPathModifier {
		b.httpRouteFilter.RequestRedirect.Path = &gatewayapi.HTTPPathModifier{
			Type:            gatewayapi.FullPathHTTPPathModifier,
			ReplaceFullPath: lo.ToPtr(path),
		}
	}
	if modifierType == gatewayapi.PrefixMatchHTTPPathModifier {
		b.httpRouteFilter.RequestRedirect.Path = &gatewayapi.HTTPPathModifier{
			Type:               gatewayapi.PrefixMatchHTTPPathModifier,
			ReplacePrefixMatch: lo.ToPtr(path),
		}
	}
	return b
}

// NewHTTPRouteRequestHeaderModifierFilter builds a request header modifier HTTPRoute filter.
func NewHTTPRouteRequestHeaderModifierFilter() *HTTPRouteFilterBuilder {
	filter := gatewayapi.HTTPRouteFilter{
		Type:                  gatewayapi.HTTPRouteFilterRequestHeaderModifier,
		RequestHeaderModifier: &gatewayapi.HTTPHeaderFilter{},
	}
	return &HTTPRouteFilterBuilder{httpRouteFilter: filter}
}

func (b *HTTPRouteFilterBuilder) WithRequestHeaderAdd(headers []gatewayapi.HTTPHeader) *HTTPRouteFilterBuilder {
	if b.httpRouteFilter.Type != gatewayapi.HTTPRouteFilterRequestHeaderModifier ||
		b.httpRouteFilter.RequestHeaderModifier == nil {
		return b
	}
	b.httpRouteFilter.RequestHeaderModifier.Add = headers
	return b
}

func (b *HTTPRouteFilterBuilder) WithRequestHeaderSet(headers []gatewayapi.HTTPHeader) *HTTPRouteFilterBuilder {
	if b.httpRouteFilter.Type != gatewayapi.HTTPRouteFilterRequestHeaderModifier ||
		b.httpRouteFilter.RequestHeaderModifier == nil {
		return b
	}
	b.httpRouteFilter.RequestHeaderModifier.Set = headers
	return b
}

func (b *HTTPRouteFilterBuilder) WithRequestHeaderRemove(headerNames []string) *HTTPRouteFilterBuilder {
	if b.httpRouteFilter.Type != gatewayapi.HTTPRouteFilterRequestHeaderModifier ||
		b.httpRouteFilter.RequestHeaderModifier == nil {
		return b
	}
	b.httpRouteFilter.RequestHeaderModifier.Remove = headerNames
	return b
}
