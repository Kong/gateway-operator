package kongstate

import (
	"fmt"

	"github.com/kong/go-kong/kong"
	"github.com/mitchellh/mapstructure"

	"github.com/kong/kong-operator/ingress-controller/internal/util"
)

// redactedString is used to redact sensitive values in the KongState.
// It uses a vault URI to pass Konnect Admin API validations (e.g. when a TLS key is expected, it's only possible
// to pass a valid key or a vault URI).
var redactedString = kong.String("{vault://redacted-value}")

// randRedactedString is used to redact sensitive values in the KongState when the value must be random to avoid
// collisions.
func randRedactedString(uuidGenerator util.UUIDGenerator) *string {
	s := fmt.Sprintf("{vault://%s}", uuidGenerator.NewString())
	return &s
}

// KeyAuth represents a key-auth credential.
type KeyAuth struct {
	kong.KeyAuth
}

// HMACAuth represents a HMAC credential.
type HMACAuth struct {
	kong.HMACAuth
}

// JWTAuth represents a JWT credential.
type JWTAuth struct {
	kong.JWTAuth
}

// BasicAuth represents a basic authentication credential.
type BasicAuth struct {
	kong.BasicAuth
}

// ACLGroup represents an ACL associated with a consumer. Due to ACL implementation in Kong being similar to
// credentials, ACLs are treated as credentials, too.
type ACLGroup struct {
	kong.ACLGroup
}

// Oauth2Credential represents an OAuth2 client configuration including credentials.
type Oauth2Credential struct {
	kong.Oauth2Credential
}

// MTLSAuth represents an MTLS auth credential.
type MTLSAuth struct {
	kong.MTLSAuth
}

func NewKeyAuth(config any) (*KeyAuth, error) {
	var res KeyAuth
	err := decodeCredential(config, &res.KeyAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key-auth credential: %w", err)
	}

	// TODO we perform these validity checks here because passing credentials without these fields will panic deck
	// later on. Ideally this should not be handled in the controller, but we cannot currently handle it elsewhere
	// (i.e. in deck or go-kong) without entering a sync failure loop that cannot actually report the problem
	// piece of configuration. if we can address those limitations, we should remove these checks.
	// See https://github.com/Kong/deck/pull/223 and https://github.com/Kong/kubernetes-ingress-controller/issues/532
	// for more discussion.
	if res.Key == nil {
		return nil, fmt.Errorf("key-auth is invalid: no key")
	}
	return &res, nil
}

func NewHMACAuth(config any) (*HMACAuth, error) {
	var res HMACAuth
	err := decodeCredential(config, &res.HMACAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hmac-auth credential: %w", err)
	}
	if res.Username == nil {
		return nil, fmt.Errorf("hmac-auth is invalid: no username")
	}
	return &res, nil
}

func NewJWTAuth(config any) (*JWTAuth, error) {
	var res JWTAuth
	err := decodeCredential(config, &res.JWTAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to process JWT credential: %w", err)
	}
	// This is treated specially because only this
	// field might be omitted by user under the expectation
	// that Kong will insert the default.
	// If we don't set it, decK will detect a diff and PUT this
	// credential everytime it performs a sync operation, which
	// leads to unnecessary cache invalidations in Kong.
	if res.Algorithm == nil || *res.Algorithm == "" {
		res.Algorithm = kong.String("HS256")
	}
	if res.Key == nil {
		return nil, fmt.Errorf("jwt-auth for is invalid: no key")
	}
	return &res, nil
}

func NewBasicAuth(config any) (*BasicAuth, error) {
	var res BasicAuth
	err := decodeCredential(config, &res.BasicAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to decode basic-auth credential: %w", err)
	}
	if res.Username == nil {
		return nil, fmt.Errorf("basic-auth is invalid: no username")
	}
	return &res, nil
}

func NewACLGroup(config any) (*ACLGroup, error) {
	var res ACLGroup
	err := decodeCredential(config, &res.ACLGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to process ACL group: %w", err)
	}
	if res.Group == nil {
		return nil, fmt.Errorf("acl is invalid: no group")
	}
	return &res, nil
}

func NewOauth2Credential(config any) (*Oauth2Credential, error) {
	var res Oauth2Credential
	err := decodeCredential(config, &res.Oauth2Credential)
	if err != nil {
		return nil, fmt.Errorf("failed to decode oauth2 credential: %w", err)
	}
	if res.ClientID == nil {
		return nil, fmt.Errorf("oauth2 is invalid: no client_id")
	}
	if res.Name == nil {
		return nil, fmt.Errorf("oauth2 is invalid: no name")
	}
	return &res, nil
}

func NewMTLSAuth(config any) (*MTLSAuth, error) {
	var res MTLSAuth
	err := decodeCredential(config, &res.MTLSAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to decode mTLS credential: %w", err)
	}
	if res.SubjectName == nil {
		return nil, fmt.Errorf("mtls-auth is invalid: no subject_name")
	}
	return &res, nil
}

// SanitizedCopy returns a shallow copy with sensitive values redacted best-effort.
func (c *KeyAuth) SanitizedCopy(uuidGenerator util.UUIDGenerator) *KeyAuth {
	return &KeyAuth{
		KeyAuth: kong.KeyAuth{
			// Consumer field omitted
			CreatedAt: c.CreatedAt,
			ID:        c.ID,
			Key:       randRedactedString(uuidGenerator),
			Tags:      c.Tags,
		},
	}
}

// SanitizedCopy returns a shallow copy with sensitive values redacted best-effort.
func (c *HMACAuth) SanitizedCopy() *HMACAuth {
	return &HMACAuth{
		HMACAuth: kong.HMACAuth{
			// Consumer field omitted
			CreatedAt: c.CreatedAt,
			ID:        c.ID,
			Username:  c.Username,
			Secret:    redactedString,
			Tags:      c.Tags,
		},
	}
}

// SanitizedCopy returns a shallow copy with sensitive values redacted best-effort.
func (c *JWTAuth) SanitizedCopy() *JWTAuth {
	return &JWTAuth{
		JWTAuth: kong.JWTAuth{
			// Consumer field omitted
			CreatedAt:    c.CreatedAt,
			ID:           c.ID,
			Algorithm:    c.Algorithm,
			Key:          c.Key, // despite field name, "key" is an identifier
			RSAPublicKey: c.RSAPublicKey,
			Secret:       redactedString,
			Tags:         c.Tags,
		},
	}
}

// SanitizedCopy returns a shallow copy with sensitive values redacted best-effort.
func (c *BasicAuth) SanitizedCopy() *BasicAuth {
	return &BasicAuth{
		BasicAuth: kong.BasicAuth{
			// Consumer field omitted
			CreatedAt: c.CreatedAt,
			ID:        c.ID,
			Username:  c.Username,
			Password:  redactedString,
			Tags:      c.Tags,
		},
	}
}

// SanitizedCopy returns a shallow copy with sensitive values redacted best-effort.
func (c *Oauth2Credential) SanitizedCopy() *Oauth2Credential {
	return &Oauth2Credential{
		Oauth2Credential: kong.Oauth2Credential{
			// Consumer field omitted
			CreatedAt:    c.CreatedAt,
			ID:           c.ID,
			Name:         c.Name,
			ClientID:     c.ClientID,
			ClientSecret: redactedString,
			RedirectURIs: c.RedirectURIs,
			Tags:         c.Tags,
		},
	}
}

func decodeCredential(credConfig any,
	credStructPointer any,
) error {
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			TagName: "json",
			Result:  credStructPointer,
		})
	if err != nil {
		return fmt.Errorf("failed to create a decoder: %w", err)
	}
	err = decoder.Decode(credConfig)
	if err != nil {
		return fmt.Errorf("failed to decode credential: %w", err)
	}
	return nil
}
