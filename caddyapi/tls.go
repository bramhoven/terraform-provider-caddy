package caddyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// re-export for use by the provider package
type RawJSON = json.RawMessage

// TLS represents the Caddy TLS app configuration
// https://caddyserver.com/docs/json/apps/tls/
type TLS struct {
	Certificates *TLSCertificates `json:"certificates,omitempty"`
	Automation   *TLSAutomation   `json:"automation,omitempty"`
}

// TLSCertificates defines certificate sources
type TLSCertificates struct {
	Automate  []string      `json:"automate,omitempty"`
	LoadFiles []TLSLoadFile `json:"load_files,omitempty"`
	LoadPEM   []TLSLoadPEM  `json:"load_pem,omitempty"`
}

// TLSLoadFile loads a certificate/key pair from disk
type TLSLoadFile struct {
	Certificate string   `json:"certificate"`
	Key         string   `json:"key"`
	Format      string   `json:"format,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// TLSLoadPEM loads a certificate/key pair from PEM-encoded strings
type TLSLoadPEM struct {
	Certificate string   `json:"certificate"`
	Key         string   `json:"key"`
	Tags        []string `json:"tags,omitempty"`
}

// TLSAutomation configures automatic certificate management
type TLSAutomation struct {
	Policies []TLSAutomationPolicy `json:"policies,omitempty"`
	OnDemand *TLSOnDemand          `json:"on_demand,omitempty"`
}

// TLSAutomationPolicy is a single automation policy
type TLSAutomationPolicy struct {
	Subjects   []string        `json:"subjects,omitempty"`
	Issuers    []IssuerMarshal `json:"issuers,omitempty"`
	KeyType    string          `json:"key_type,omitempty"`
	MustStaple bool            `json:"must_staple,omitempty"`
	OnDemand   bool            `json:"on_demand,omitempty"`
}

// TLSOnDemand configures on-demand TLS
type TLSOnDemand struct {
	Ask       string           `json:"ask,omitempty"`
	RateLimit *TLSOnDemandRate `json:"rate_limit,omitempty"`
}

// TLSOnDemandRate rate-limits on-demand certificate issuance
type TLSOnDemandRate struct {
	Interval Duration `json:"interval,omitempty"`
	Burst    int      `json:"burst,omitempty"`
}

// ACMEIssuer represents an ACME certificate issuer
type ACMEIssuer struct {
	CA                   string               `json:"ca,omitempty"`
	Email                string               `json:"email,omitempty"`
	ExternalAccount      *ACMEExternalAccount `json:"external_account,omitempty"`
	Challenges           *ACMEChallenges      `json:"challenges,omitempty"`
	TrustedRootsPEMFiles []string             `json:"trusted_roots_pem_files,omitempty"`
}

// ACMEExternalAccount holds External Account Binding (EAB) credentials
type ACMEExternalAccount struct {
	KeyID  string `json:"key_id"`
	MACKey string `json:"mac_key"`
}

// ACMEChallenges configures which ACME challenge types to use
type ACMEChallenges struct {
	HTTP    *ACMEChallenge `json:"http,omitempty"`
	TLSALPN *ACMEChallenge `json:"tls-alpn,omitempty"`
	DNS     *DNSChallenge  `json:"dns,omitempty"`
}

// ACMEChallenge configures a single ACME challenge type (HTTP-01 or TLS-ALPN-01)
type ACMEChallenge struct {
	Disabled      bool `json:"disabled,omitempty"`
	AlternatePort int  `json:"alternate_port,omitempty"`
}

// DNSChallenge configures the DNS-01 ACME challenge
// https://caddyserver.com/docs/json/apps/tls/automation/policies/issuers/acme/challenges/dns/
type DNSChallenge struct {
	// Provider is the DNS provider module config. JSON object with a "name" key
	// identifying the provider (e.g. "cloudflare") and provider-specific fields.
	Provider           RawJSON  `json:"provider,omitempty"`
	TTL                Duration `json:"ttl,omitempty"`
	PropagationDelay   Duration `json:"propagation_delay,omitempty"`
	PropagationTimeout Duration `json:"propagation_timeout,omitempty"`
	Resolvers          []string `json:"resolvers,omitempty"`
	OverrideDomain     string   `json:"override_domain,omitempty"`
}

// InternalIssuer represents an internal (self-signed) certificate issuer
type InternalIssuer struct {
	CA           string   `json:"ca,omitempty"`
	Lifetime     Duration `json:"lifetime,omitempty"`
	SignWithRoot bool     `json:"sign_with_root,omitempty"`
}

// ZeroSSLIssuer represents a ZeroSSL certificate issuer
type ZeroSSLIssuer struct {
	CA              string               `json:"ca,omitempty"`
	Email           string               `json:"email,omitempty"`
	APIKey          string               `json:"api_key,omitempty"`
	ExternalAccount *ACMEExternalAccount `json:"external_account,omitempty"`
}

// IssuerMarshal handles polymorphic JSON marshaling/unmarshaling for TLS issuers.
// Issuers use a "module" field to identify the type, similar to how handlers use "handler".
type IssuerMarshal struct {
	Issuer interface{}
}

// ToIssuerModule returns the Caddy module name for an issuer type
func ToIssuerModule(i interface{}) string {
	switch i.(type) {
	case ACMEIssuer:
		return "acme"
	case InternalIssuer:
		return "internal"
	case ZeroSSLIssuer:
		return "zerossl"
	default:
		return ""
	}
}

// FromModuleName creates the appropriate issuer from a module name and JSON bytes
func (s *IssuerMarshal) FromModuleName(name string, b []byte) error {
	switch name {
	case "acme":
		issuer := ACMEIssuer{}
		if err := json.Unmarshal(b, &issuer); err != nil {
			return err
		}
		s.Issuer = issuer
	case "internal":
		issuer := InternalIssuer{}
		if err := json.Unmarshal(b, &issuer); err != nil {
			return err
		}
		s.Issuer = issuer
	case "zerossl":
		issuer := ZeroSSLIssuer{}
		if err := json.Unmarshal(b, &issuer); err != nil {
			return err
		}
		s.Issuer = issuer
	default:
		return fmt.Errorf("unsupported issuer module %s", name)
	}
	return nil
}

// MarshalJSON adds the "module" field to identify the issuer type
func (s IssuerMarshal) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(s.Issuer)
	if err != nil {
		return nil, err
	}
	issuer := make(map[string]interface{})
	if err := json.Unmarshal(b, &issuer); err != nil {
		return nil, err
	}
	issuer["module"] = ToIssuerModule(s.Issuer)
	if issuer["module"] == "" {
		return nil, fmt.Errorf("unsupported issuer %v", s.Issuer)
	}
	return json.Marshal(issuer)
}

// UnmarshalJSON reads the "module" field to determine the issuer type
func (s *IssuerMarshal) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	if err := json.Unmarshal(b, &object); err != nil {
		return err
	}
	var module string
	m, ok := object["module"]
	if !ok {
		return fmt.Errorf("missing required field 'module'")
	}
	if err := json.Unmarshal([]byte(m), &module); err != nil {
		return err
	}

	delete(object, "module")
	b, err := json.Marshal(object)
	if err != nil {
		return err
	}

	return s.FromModuleName(module, b)
}

// CreateTLS creates the TLS app configuration.
// Uses POST which creates or replaces (PUT would 409 if the key already exists).
func (c *Client) CreateTLS(tls TLS) error {
	if err := c.EnforceExists("@config/apps"); err != nil {
		return fmt.Errorf("CreateTLS: %w", err)
	}

	resp, err := c.client.R().SetBody(tls).Post("/config/apps/tls")
	if err != nil {
		return fmt.Errorf("CreateTLS: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("CreateTLS: %w", StatusError{resp})
	}
	return nil
}

// GetTLS retrieves the current TLS app configuration
func (c *Client) GetTLS() (*TLS, error) {
	resp, err := c.client.R().SetResult(&TLS{}).Get("/config/apps/tls")
	if err != nil {
		return nil, fmt.Errorf("GetTLS: %w", err)
	}
	if resp.IsError() {
		if resp.StatusCode() == http.StatusNotFound || strings.Contains(string(resp.Body()), "invalid traversal path") {
			return nil, fmt.Errorf("GetTLS: %w", ErrConfigPathNotFound)
		}
		return nil, fmt.Errorf("GetTLS: %w", StatusError{resp})
	}
	return resp.Result().(*TLS), nil
}

// UpdateTLS replaces the TLS app configuration.
// Uses PATCH which replaces an existing value (PUT would 409 if it already exists).
func (c *Client) UpdateTLS(tls TLS) error {
	resp, err := c.client.R().SetBody(tls).Patch("/config/apps/tls")
	if err != nil {
		return fmt.Errorf("UpdateTLS: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("UpdateTLS: %w", StatusError{resp})
	}
	return nil
}

// DeleteTLS removes the TLS app configuration
func (c *Client) DeleteTLS() error {
	resp, err := c.client.R().Delete("/config/apps/tls")
	if err != nil {
		return fmt.Errorf("DeleteTLS: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("DeleteTLS: %w", StatusError{resp})
	}
	return nil
}
