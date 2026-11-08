package caddyapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuerMarshalACME(t *testing.T) {
	issuer := IssuerMarshal{
		Issuer: ACMEIssuer{
			CA:    "https://acme-v02.api.letsencrypt.org/directory",
			Email: "admin@example.com",
		},
	}

	b, err := json.Marshal(issuer)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "acme", m["module"])
	assert.Equal(t, "https://acme-v02.api.letsencrypt.org/directory", m["ca"])
	assert.Equal(t, "admin@example.com", m["email"])

	var decoded IssuerMarshal
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, issuer, decoded)
}

func TestIssuerMarshalInternal(t *testing.T) {
	issuer := IssuerMarshal{
		Issuer: InternalIssuer{
			CA:           "local",
			SignWithRoot: true,
		},
	}

	b, err := json.Marshal(issuer)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "internal", m["module"])
	assert.Equal(t, "local", m["ca"])
	assert.Equal(t, true, m["sign_with_root"])

	var decoded IssuerMarshal
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, issuer, decoded)
}

func TestIssuerMarshalZeroSSL(t *testing.T) {
	issuer := IssuerMarshal{
		Issuer: ZeroSSLIssuer{
			Email:  "admin@example.com",
			APIKey: "test-key",
		},
	}

	b, err := json.Marshal(issuer)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "zerossl", m["module"])
	assert.Equal(t, "admin@example.com", m["email"])
	assert.Equal(t, "test-key", m["api_key"])

	var decoded IssuerMarshal
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, issuer, decoded)
}

func TestTLSMarshalFull(t *testing.T) {
	tls := TLS{
		Certificates: &TLSCertificates{
			Automate: []string{"example.com", "www.example.com"},
			LoadFiles: []TLSLoadFile{
				{
					Certificate: "/path/cert.pem",
					Key:         "/path/key.pem",
					Format:      "pem",
				},
			},
		},
		Automation: &TLSAutomation{
			Policies: []TLSAutomationPolicy{
				{
					Subjects: []string{"example.com"},
					KeyType:  "ec256",
					Issuers: []IssuerMarshal{
						{Issuer: ACMEIssuer{
							CA:    "https://acme-v02.api.letsencrypt.org/directory",
							Email: "admin@example.com",
							Challenges: &ACMEChallenges{
								HTTP: &ACMEChallenge{AlternatePort: 8080},
							},
						}},
					},
				},
			},
			OnDemand: &TLSOnDemand{
				Ask: "https://auth.example.com/check",
			},
		},
	}

	b, err := json.Marshal(tls)
	require.NoError(t, err)

	var decoded TLS
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, tls, decoded)
}

func TestACMEIssuerWithEAB(t *testing.T) {
	issuer := IssuerMarshal{
		Issuer: ACMEIssuer{
			CA:    "https://acme.zerossl.com/v2/DV90",
			Email: "admin@example.com",
			ExternalAccount: &ACMEExternalAccount{
				KeyID:  "kid123",
				MACKey: "mac456",
			},
		},
	}

	b, err := json.Marshal(issuer)
	require.NoError(t, err)

	var decoded IssuerMarshal
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, issuer, decoded)

	acme := decoded.Issuer.(ACMEIssuer)
	assert.Equal(t, "kid123", acme.ExternalAccount.KeyID)
	assert.Equal(t, "mac456", acme.ExternalAccount.MACKey)
}
