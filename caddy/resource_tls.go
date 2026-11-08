package caddy

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/conradludgate/terraform-provider-caddy/caddyapi"
	"github.com/conradludgate/tfutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TLSApp represents the caddy TLS Application resource
type TLSApp struct{}

// TLSIssuer defines the schema for a polymorphic TLS issuer (acme, internal, zerossl)
type TLSIssuer struct{}

func (TLSIssuer) Schema() tfutils.SchemaMap {
	return tfutils.SchemaMap{
		"acme": tfutils.SchemaMap{
			"ca":    tfutils.String().Optional(true),
			"email": tfutils.String().Optional(true),
			"external_account": tfutils.SchemaMap{
				"key_id":  tfutils.String().Required(true),
				"mac_key": tfutils.String().Required(true),
			}.IntoSet().MaxItems(1).Optional(true),
			"challenges": tfutils.SchemaMap{
				"http": tfutils.SchemaMap{
					"disabled":       tfutils.Bool().Optional(true),
					"alternate_port": tfutils.Int().Optional(true),
				}.IntoSet().MaxItems(1).Optional(true),
				"tls_alpn": tfutils.SchemaMap{
					"disabled":       tfutils.Bool().Optional(true),
					"alternate_port": tfutils.Int().Optional(true),
				}.IntoSet().MaxItems(1).Optional(true),
				"dns": tfutils.SchemaMap{
					"provider":             tfutils.String().Required(true),
					"ttl":                  tfutils.String().Optional(true),
					"propagation_delay":    tfutils.String().Optional(true),
					"propagation_timeout":  tfutils.String().Optional(true),
					"resolvers":            tfutils.String().List().Optional(true),
					"override_domain":      tfutils.String().Optional(true),
				}.IntoSet().MaxItems(1).Optional(true),
			}.IntoSet().MaxItems(1).Optional(true),
			"trusted_roots_pem_files": tfutils.String().List().Optional(true),
		}.IntoSet().MaxItems(1).Optional(true),

		"internal": tfutils.SchemaMap{
			"ca":             tfutils.String().Optional(true),
			"lifetime":       tfutils.String().Optional(true),
			"sign_with_root": tfutils.Bool().Optional(true),
		}.IntoSet().MaxItems(1).Optional(true),

		"zerossl": tfutils.SchemaMap{
			"ca":      tfutils.String().Optional(true),
			"email":   tfutils.String().Optional(true),
			"api_key": tfutils.String().Optional(true),
			"external_account": tfutils.SchemaMap{
				"key_id":  tfutils.String().Required(true),
				"mac_key": tfutils.String().Required(true),
			}.IntoSet().MaxItems(1).Optional(true),
		}.IntoSet().MaxItems(1).Optional(true),
	}
}

// TLSPolicy defines the schema for an automation policy
type TLSPolicy struct{}

func (TLSPolicy) Schema() tfutils.SchemaMap {
	return tfutils.SchemaMap{
		"subjects":    tfutils.String().List().Optional(true),
		"key_type":    tfutils.String().Optional(true),
		"must_staple": tfutils.Bool().Optional(true),
		"on_demand":   tfutils.Bool().Optional(true),
		"issuer":      tfutils.ListOf(TLSIssuer{}).Optional(true),
	}
}

func (TLSApp) Schema() tfutils.SchemaMap {
	return tfutils.SchemaMap{
		"certificates": tfutils.SchemaMap{
			"automate": tfutils.String().List().Optional(true),
			"load_files": tfutils.SchemaMap{
				"certificate": tfutils.String().Required(true),
				"key":         tfutils.String().Required(true),
				"format":      tfutils.String().Optional(true),
				"tags":        tfutils.String().List().Optional(true),
			}.IntoList().Optional(true),
			"load_pem": tfutils.SchemaMap{
				"certificate": tfutils.String().Required(true),
				"key":         tfutils.String().Required(true),
				"tags":        tfutils.String().List().Optional(true),
			}.IntoList().Optional(true),
		}.IntoSet().MaxItems(1).Optional(true),

		"automation": tfutils.SchemaMap{
			"policy": tfutils.ListOf(TLSPolicy{}).Optional(true),
			"on_demand": tfutils.SchemaMap{
				"ask": tfutils.String().Optional(true),
				"rate_limit": tfutils.SchemaMap{
					"interval": tfutils.String().Optional(true),
					"burst":    tfutils.Int().Optional(true),
				}.IntoSet().MaxItems(1).Optional(true),
			}.IntoSet().MaxItems(1).Optional(true),
		}.IntoSet().MaxItems(1).Optional(true),
	}
}

func (TLSApp) Create(d *schema.ResourceData, m interface{}) error {
	c := m.(Client)

	tls := tlsFromResourceData(d)
	if err := c.CreateTLS(tls); err != nil {
		return err
	}
	d.SetId("@config/apps/tls")

	return nil
}

func (TLSApp) Read(d *schema.ResourceData, m interface{}) error {
	c := m.(Client)

	if d.Id() == "" {
		return nil
	}

	tls, err := c.GetTLS()
	if err != nil {
		if errors.Is(err, caddyapi.ErrConfigPathNotFound) || strings.Contains(err.Error(), "invalid traversal path") {
			d.SetId("")
			return nil
		}
		return err
	}

	if tls.Certificates != nil {
		d.Set("certificates", tlsCertificatesInto(tls.Certificates))
	}

	if tls.Automation != nil {
		d.Set("automation", tlsAutomationInto(tls.Automation))
	}

	return nil
}

func (TLSApp) Update(d *schema.ResourceData, m interface{}) error {
	c := m.(Client)
	tls := tlsFromResourceData(d)
	return c.UpdateTLS(tls)
}

func (TLSApp) Delete(d *schema.ResourceData, m interface{}) error {
	c := m.(Client)
	return c.DeleteTLS()
}

// --- Terraform → Caddy conversion helpers ---

func tlsFromResourceData(d *schema.ResourceData) caddyapi.TLS {
	tls := caddyapi.TLS{}

	if certs := GetObjectSet(d, "certificates"); len(certs) == 1 {
		tls.Certificates = tlsCertificatesFrom(&certs[0])
	}

	if auto := GetObjectSet(d, "automation"); len(auto) == 1 {
		tls.Automation = tlsAutomationFrom(&auto[0])
	}

	return tls
}

func tlsCertificatesFrom(d *MapData) *caddyapi.TLSCertificates {
	certs := &caddyapi.TLSCertificates{
		Automate: GetStringList(d, "automate"),
	}

	if files := GetObjectList(d, "load_files"); len(files) > 0 {
		certs.LoadFiles = make([]caddyapi.TLSLoadFile, len(files))
		for i, f := range files {
			certs.LoadFiles[i] = caddyapi.TLSLoadFile{
				Certificate: GetString(&f, "certificate"),
				Key:         GetString(&f, "key"),
				Format:      GetString(&f, "format"),
				Tags:        GetStringList(&f, "tags"),
			}
		}
	}

	if pems := GetObjectList(d, "load_pem"); len(pems) > 0 {
		certs.LoadPEM = make([]caddyapi.TLSLoadPEM, len(pems))
		for i, p := range pems {
			certs.LoadPEM[i] = caddyapi.TLSLoadPEM{
				Certificate: GetString(&p, "certificate"),
				Key:         GetString(&p, "key"),
				Tags:        GetStringList(&p, "tags"),
			}
		}
	}

	return certs
}

func tlsAutomationFrom(d *MapData) *caddyapi.TLSAutomation {
	auto := &caddyapi.TLSAutomation{}

	if policies := GetObjectList(d, "policy"); len(policies) > 0 {
		auto.Policies = make([]caddyapi.TLSAutomationPolicy, len(policies))
		for i, p := range policies {
			auto.Policies[i] = tlsAutomationPolicyFrom(&p)
		}
	}

	if od := GetObjectSet(d, "on_demand"); len(od) == 1 {
		auto.OnDemand = tlsOnDemandFrom(&od[0])
	}

	return auto
}

func tlsAutomationPolicyFrom(d *MapData) caddyapi.TLSAutomationPolicy {
	policy := caddyapi.TLSAutomationPolicy{
		Subjects:   GetStringList(d, "subjects"),
		KeyType:    GetString(d, "key_type"),
		MustStaple: GetBool(d, "must_staple"),
		OnDemand:   GetBool(d, "on_demand"),
	}

	if issuers := GetObjectList(d, "issuer"); len(issuers) > 0 {
		policy.Issuers = tlsIssuersFrom(issuers)
	}

	return policy
}

func tlsIssuersFrom(issuers []MapData) []caddyapi.IssuerMarshal {
	result := make([]caddyapi.IssuerMarshal, 0, len(issuers))
	for _, d := range issuers {
		result = append(result, tlsIssuerFrom(&d))
	}
	return result
}

func tlsIssuerFrom(d *MapData) caddyapi.IssuerMarshal {
	for key, v := range *d {
		s := AsSet(v)
		if len(s) != 1 {
			continue
		}
		d := &s[0]

		switch key {
		case "acme":
			return caddyapi.IssuerMarshal{Issuer: tlsACMEIssuerFrom(d)}
		case "internal":
			return caddyapi.IssuerMarshal{Issuer: tlsInternalIssuerFrom(d)}
		case "zerossl":
			return caddyapi.IssuerMarshal{Issuer: tlsZeroSSLIssuerFrom(d)}
		}
	}

	panic("no issuer type specified")
}

func tlsACMEIssuerFrom(d *MapData) caddyapi.ACMEIssuer {
	issuer := caddyapi.ACMEIssuer{
		CA:                   GetString(d, "ca"),
		Email:                GetString(d, "email"),
		TrustedRootsPEMFiles: GetStringList(d, "trusted_roots_pem_files"),
	}

	if ea := GetObjectSet(d, "external_account"); len(ea) == 1 {
		issuer.ExternalAccount = &caddyapi.ACMEExternalAccount{
			KeyID:  GetString(&ea[0], "key_id"),
			MACKey: GetString(&ea[0], "mac_key"),
		}
	}

	if ch := GetObjectSet(d, "challenges"); len(ch) == 1 {
		issuer.Challenges = &caddyapi.ACMEChallenges{}
		if http := GetObjectSet(&ch[0], "http"); len(http) == 1 {
			issuer.Challenges.HTTP = &caddyapi.ACMEChallenge{
				Disabled:      GetBool(&http[0], "disabled"),
				AlternatePort: GetInt(&http[0], "alternate_port"),
			}
		}
		if tlsAlpn := GetObjectSet(&ch[0], "tls_alpn"); len(tlsAlpn) == 1 {
			issuer.Challenges.TLSALPN = &caddyapi.ACMEChallenge{
				Disabled:      GetBool(&tlsAlpn[0], "disabled"),
				AlternatePort: GetInt(&tlsAlpn[0], "alternate_port"),
			}
		}
		if dns := GetObjectSet(&ch[0], "dns"); len(dns) == 1 {
			issuer.Challenges.DNS = &caddyapi.DNSChallenge{
				Resolvers:      GetStringList(&dns[0], "resolvers"),
				OverrideDomain: GetString(&dns[0], "override_domain"),
			}
			if provider := GetString(&dns[0], "provider"); provider != "" {
				issuer.Challenges.DNS.Provider = json.RawMessage(provider)
			}
			issuer.Challenges.DNS.TTL.UnmarshalText(GetString(&dns[0], "ttl"))
			issuer.Challenges.DNS.PropagationDelay.UnmarshalText(GetString(&dns[0], "propagation_delay"))
			issuer.Challenges.DNS.PropagationTimeout.UnmarshalText(GetString(&dns[0], "propagation_timeout"))
		}
	}

	return issuer
}

func tlsInternalIssuerFrom(d *MapData) caddyapi.InternalIssuer {
	var lifetime caddyapi.Duration
	lifetime.UnmarshalText(GetString(d, "lifetime"))

	return caddyapi.InternalIssuer{
		CA:           GetString(d, "ca"),
		Lifetime:     lifetime,
		SignWithRoot: GetBool(d, "sign_with_root"),
	}
}

func tlsZeroSSLIssuerFrom(d *MapData) caddyapi.ZeroSSLIssuer {
	issuer := caddyapi.ZeroSSLIssuer{
		CA:     GetString(d, "ca"),
		Email:  GetString(d, "email"),
		APIKey: GetString(d, "api_key"),
	}

	if ea := GetObjectSet(d, "external_account"); len(ea) == 1 {
		issuer.ExternalAccount = &caddyapi.ACMEExternalAccount{
			KeyID:  GetString(&ea[0], "key_id"),
			MACKey: GetString(&ea[0], "mac_key"),
		}
	}

	return issuer
}

func tlsOnDemandFrom(d *MapData) *caddyapi.TLSOnDemand {
	od := &caddyapi.TLSOnDemand{
		Ask: GetString(d, "ask"),
	}

	if rl := GetObjectSet(d, "rate_limit"); len(rl) == 1 {
		od.RateLimit = &caddyapi.TLSOnDemandRate{
			Burst: GetInt(&rl[0], "burst"),
		}
		od.RateLimit.Interval.UnmarshalText(GetString(&rl[0], "interval"))
	}

	return od
}

// --- Caddy → Terraform state conversion helpers ---

func tlsCertificatesInto(certs *caddyapi.TLSCertificates) []interface{} {
	m := map[string]interface{}{}

	if certs.Automate != nil {
		automate := make([]interface{}, len(certs.Automate))
		for i, v := range certs.Automate {
			automate[i] = v
		}
		m["automate"] = automate
	}

	if certs.LoadFiles != nil {
		files := make([]interface{}, len(certs.LoadFiles))
		for i, f := range certs.LoadFiles {
			file := map[string]interface{}{
				"certificate": f.Certificate,
				"key":         f.Key,
				"format":      f.Format,
			}
			if f.Tags != nil {
				tags := make([]interface{}, len(f.Tags))
				for j, t := range f.Tags {
					tags[j] = t
				}
				file["tags"] = tags
			}
			files[i] = file
		}
		m["load_files"] = files
	}

	if certs.LoadPEM != nil {
		pems := make([]interface{}, len(certs.LoadPEM))
		for i, p := range certs.LoadPEM {
			pem := map[string]interface{}{
				"certificate": p.Certificate,
				"key":         p.Key,
			}
			if p.Tags != nil {
				tags := make([]interface{}, len(p.Tags))
				for j, t := range p.Tags {
					tags[j] = t
				}
				pem["tags"] = tags
			}
			pems[i] = pem
		}
		m["load_pem"] = pems
	}

	return []interface{}{m}
}

func tlsAutomationInto(auto *caddyapi.TLSAutomation) []interface{} {
	m := map[string]interface{}{}

	if auto.Policies != nil {
		policies := make([]interface{}, len(auto.Policies))
		for i, p := range auto.Policies {
			policies[i] = tlsAutomationPolicyInto(p)
		}
		m["policy"] = policies
	}

	if auto.OnDemand != nil {
		m["on_demand"] = tlsOnDemandInto(auto.OnDemand)
	}

	return []interface{}{m}
}

func tlsAutomationPolicyInto(p caddyapi.TLSAutomationPolicy) map[string]interface{} {
	m := map[string]interface{}{
		"key_type":    p.KeyType,
		"must_staple": p.MustStaple,
		"on_demand":   p.OnDemand,
	}

	if p.Subjects != nil {
		subjects := make([]interface{}, len(p.Subjects))
		for i, s := range p.Subjects {
			subjects[i] = s
		}
		m["subjects"] = subjects
	}

	if p.Issuers != nil {
		issuers := make([]interface{}, len(p.Issuers))
		for i, issuer := range p.Issuers {
			issuers[i] = tlsIssuerInto(issuer)
		}
		m["issuer"] = issuers
	}

	return m
}

func tlsIssuerInto(issuer caddyapi.IssuerMarshal) map[string]interface{} {
	m := map[string]interface{}{}

	switch i := issuer.Issuer.(type) {
	case caddyapi.ACMEIssuer:
		m["acme"] = schema.NewSet(schema.HashResource(
			TLSIssuer{}.Schema().BuildResource().Schema["acme"].Elem.(*schema.Resource),
		), []interface{}{fromACMEIssuer(i)})
	case caddyapi.InternalIssuer:
		m["internal"] = schema.NewSet(schema.HashResource(
			TLSIssuer{}.Schema().BuildResource().Schema["internal"].Elem.(*schema.Resource),
		), []interface{}{fromInternalIssuer(i)})
	case caddyapi.ZeroSSLIssuer:
		m["zerossl"] = schema.NewSet(schema.HashResource(
			TLSIssuer{}.Schema().BuildResource().Schema["zerossl"].Elem.(*schema.Resource),
		), []interface{}{fromZeroSSLIssuer(i)})
	}

	return m
}

func fromACMEIssuer(i caddyapi.ACMEIssuer) map[string]interface{} {
	m := map[string]interface{}{
		"ca":    i.CA,
		"email": i.Email,
	}

	if i.TrustedRootsPEMFiles != nil {
		roots := make([]interface{}, len(i.TrustedRootsPEMFiles))
		for j, r := range i.TrustedRootsPEMFiles {
			roots[j] = r
		}
		m["trusted_roots_pem_files"] = roots
	}

	if i.ExternalAccount != nil {
		m["external_account"] = schema.NewSet(schema.HashResource(
			TLSIssuer{}.Schema().BuildResource().Schema["acme"].Elem.(*schema.Resource).Schema["external_account"].Elem.(*schema.Resource),
		), []interface{}{map[string]interface{}{
			"key_id":  i.ExternalAccount.KeyID,
			"mac_key": i.ExternalAccount.MACKey,
		}})
	}

	if i.Challenges != nil {
		challenges := map[string]interface{}{}
		acmeSchema := TLSIssuer{}.Schema().BuildResource().Schema["acme"].Elem.(*schema.Resource).Schema["challenges"].Elem.(*schema.Resource)

		if i.Challenges.HTTP != nil {
			challenges["http"] = schema.NewSet(schema.HashResource(
				acmeSchema.Schema["http"].Elem.(*schema.Resource),
			), []interface{}{map[string]interface{}{
				"disabled":       i.Challenges.HTTP.Disabled,
				"alternate_port": i.Challenges.HTTP.AlternatePort,
			}})
		}
		if i.Challenges.TLSALPN != nil {
			challenges["tls_alpn"] = schema.NewSet(schema.HashResource(
				acmeSchema.Schema["tls_alpn"].Elem.(*schema.Resource),
			), []interface{}{map[string]interface{}{
				"disabled":       i.Challenges.TLSALPN.Disabled,
				"alternate_port": i.Challenges.TLSALPN.AlternatePort,
			}})
		}
		if i.Challenges.DNS != nil {
			dnsMap := map[string]interface{}{
				"provider":             string(i.Challenges.DNS.Provider),
				"ttl":                  i.Challenges.DNS.TTL.String(),
				"propagation_delay":    i.Challenges.DNS.PropagationDelay.String(),
				"propagation_timeout":  i.Challenges.DNS.PropagationTimeout.String(),
				"override_domain":      i.Challenges.DNS.OverrideDomain,
			}
			if i.Challenges.DNS.Resolvers != nil {
				resolvers := make([]interface{}, len(i.Challenges.DNS.Resolvers))
				for j, r := range i.Challenges.DNS.Resolvers {
					resolvers[j] = r
				}
				dnsMap["resolvers"] = resolvers
			}
			challenges["dns"] = schema.NewSet(schema.HashResource(
				acmeSchema.Schema["dns"].Elem.(*schema.Resource),
			), []interface{}{dnsMap})
		}

		m["challenges"] = schema.NewSet(schema.HashResource(acmeSchema), []interface{}{challenges})
	}

	return m
}

func fromInternalIssuer(i caddyapi.InternalIssuer) map[string]interface{} {
	return map[string]interface{}{
		"ca":             i.CA,
		"lifetime":       i.Lifetime.String(),
		"sign_with_root": i.SignWithRoot,
	}
}

func fromZeroSSLIssuer(i caddyapi.ZeroSSLIssuer) map[string]interface{} {
	m := map[string]interface{}{
		"ca":      i.CA,
		"email":   i.Email,
		"api_key": i.APIKey,
	}

	if i.ExternalAccount != nil {
		m["external_account"] = schema.NewSet(schema.HashResource(
			TLSIssuer{}.Schema().BuildResource().Schema["zerossl"].Elem.(*schema.Resource).Schema["external_account"].Elem.(*schema.Resource),
		), []interface{}{map[string]interface{}{
			"key_id":  i.ExternalAccount.KeyID,
			"mac_key": i.ExternalAccount.MACKey,
		}})
	}

	return m
}

func tlsOnDemandInto(od *caddyapi.TLSOnDemand) []interface{} {
	m := map[string]interface{}{
		"ask": od.Ask,
	}

	if od.RateLimit != nil {
		m["rate_limit"] = schema.NewSet(schema.HashResource(
			TLSApp{}.Schema().BuildResource().Schema["automation"].Elem.(*schema.Resource).Schema["on_demand"].Elem.(*schema.Resource).Schema["rate_limit"].Elem.(*schema.Resource),
		), []interface{}{map[string]interface{}{
			"interval": od.RateLimit.Interval.String(),
			"burst":    od.RateLimit.Burst,
		}})
	}

	return []interface{}{m}
}
