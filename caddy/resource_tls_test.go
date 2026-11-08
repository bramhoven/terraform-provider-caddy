package caddy

import (
	"testing"

	"github.com/conradludgate/terraform-provider-caddy/caddy/mocks"
	"github.com/conradludgate/terraform-provider-caddy/caddyapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestTLSAutomate(t *testing.T) {
	UnitTest(t, func(caddyMock *mocks.Client) {
		caddyMock.On("CreateTLS", tlsAutomate).Return(nil)
		caddyMock.On("GetTLS").Return(&tlsAutomate, nil)
		caddyMock.On("DeleteTLS").Return(nil)
	},
		resource.TestStep{
			Config: tlsAutomateConfig,
		},
	)
}

const tlsAutomateConfig = `
resource "caddy_tls" "tls" {
	certificates {
		automate = ["example.com", "www.example.com"]
	}
}
`

var tlsAutomate = caddyapi.TLS{
	Certificates: &caddyapi.TLSCertificates{
		Automate: []string{"example.com", "www.example.com"},
	},
}

func TestTLSACMEIssuer(t *testing.T) {
	UnitTest(t, func(caddyMock *mocks.Client) {
		caddyMock.On("CreateTLS", tlsACME).Return(nil)
		caddyMock.On("GetTLS").Return(&tlsACME, nil)
		caddyMock.On("DeleteTLS").Return(nil)
	},
		resource.TestStep{
			Config: tlsACMEConfig,
		},
	)
}

const tlsACMEConfig = `
resource "caddy_tls" "tls" {
	automation {
		policy {
			subjects = ["example.com"]

			issuer {
				acme {
					ca    = "https://acme-v02.api.letsencrypt.org/directory"
					email = "admin@example.com"
				}
			}
		}
	}
}
`

var tlsACME = caddyapi.TLS{
	Automation: &caddyapi.TLSAutomation{
		Policies: []caddyapi.TLSAutomationPolicy{
			{
				Subjects: []string{"example.com"},
				Issuers: []caddyapi.IssuerMarshal{
					{
						Issuer: caddyapi.ACMEIssuer{
							CA:    "https://acme-v02.api.letsencrypt.org/directory",
							Email: "admin@example.com",
						},
					},
				},
			},
		},
	},
}

func TestTLSInternalIssuer(t *testing.T) {
	UnitTest(t, func(caddyMock *mocks.Client) {
		caddyMock.On("CreateTLS", tlsInternal).Return(nil)
		caddyMock.On("GetTLS").Return(&tlsInternal, nil)
		caddyMock.On("DeleteTLS").Return(nil)
	},
		resource.TestStep{
			Config: tlsInternalConfig,
		},
	)
}

const tlsInternalConfig = `
resource "caddy_tls" "tls" {
	automation {
		policy {
			subjects = ["internal.example.com"]

			issuer {
				internal {
					ca = "local"
				}
			}
		}
	}
}
`

var tlsInternal = caddyapi.TLS{
	Automation: &caddyapi.TLSAutomation{
		Policies: []caddyapi.TLSAutomationPolicy{
			{
				Subjects: []string{"internal.example.com"},
				Issuers: []caddyapi.IssuerMarshal{
					{
						Issuer: caddyapi.InternalIssuer{
							CA: "local",
						},
					},
				},
			},
		},
	},
}

func TestTLSLoadFiles(t *testing.T) {
	UnitTest(t, func(caddyMock *mocks.Client) {
		caddyMock.On("CreateTLS", tlsLoadFiles).Return(nil)
		caddyMock.On("GetTLS").Return(&tlsLoadFiles, nil)
		caddyMock.On("DeleteTLS").Return(nil)
	},
		resource.TestStep{
			Config: tlsLoadFilesConfig,
		},
	)
}

const tlsLoadFilesConfig = `
resource "caddy_tls" "tls" {
	certificates {
		load_files {
			certificate = "/etc/ssl/certs/example.pem"
			key         = "/etc/ssl/private/example.key"
			format      = "pem"
		}
	}
}
`

var tlsLoadFiles = caddyapi.TLS{
	Certificates: &caddyapi.TLSCertificates{
		LoadFiles: []caddyapi.TLSLoadFile{
			{
				Certificate: "/etc/ssl/certs/example.pem",
				Key:         "/etc/ssl/private/example.key",
				Format:      "pem",
			},
		},
	},
}

func TestTLSACMEWithChallenges(t *testing.T) {
	UnitTest(t, func(caddyMock *mocks.Client) {
		caddyMock.On("CreateTLS", tlsACMEChallenges).Return(nil)
		caddyMock.On("GetTLS").Return(&tlsACMEChallenges, nil)
		caddyMock.On("DeleteTLS").Return(nil)
	},
		resource.TestStep{
			Config: tlsACMEChallengesConfig,
		},
	)
}

const tlsACMEChallengesConfig = `
resource "caddy_tls" "tls" {
	automation {
		policy {
			subjects = ["example.com"]
			key_type = "ec256"

			issuer {
				acme {
					ca    = "https://acme-v02.api.letsencrypt.org/directory"
					email = "admin@example.com"

					challenges {
						http {
							alternate_port = 8080
						}
						tls_alpn {
							alternate_port = 8443
						}
					}
				}
			}
		}
	}
}
`

var tlsACMEChallenges = caddyapi.TLS{
	Automation: &caddyapi.TLSAutomation{
		Policies: []caddyapi.TLSAutomationPolicy{
			{
				Subjects: []string{"example.com"},
				KeyType:  "ec256",
				Issuers: []caddyapi.IssuerMarshal{
					{
						Issuer: caddyapi.ACMEIssuer{
							CA:    "https://acme-v02.api.letsencrypt.org/directory",
							Email: "admin@example.com",
							Challenges: &caddyapi.ACMEChallenges{
								HTTP: &caddyapi.ACMEChallenge{
									AlternatePort: 8080,
								},
								TLSALPN: &caddyapi.ACMEChallenge{
									AlternatePort: 8443,
								},
							},
						},
					},
				},
			},
		},
	},
}

func TestTLSOnDemand(t *testing.T) {
	UnitTest(t, func(caddyMock *mocks.Client) {
		caddyMock.On("CreateTLS", tlsOnDemand).Return(nil)
		caddyMock.On("GetTLS").Return(&tlsOnDemand, nil)
		caddyMock.On("DeleteTLS").Return(nil)
	},
		resource.TestStep{
			Config: tlsOnDemandConfig,
		},
	)
}

const tlsOnDemandConfig = `
resource "caddy_tls" "tls" {
	automation {
		on_demand {
			ask = "https://auth.example.com/check"
		}
	}
}
`

var tlsOnDemand = caddyapi.TLS{
	Automation: &caddyapi.TLSAutomation{
		OnDemand: &caddyapi.TLSOnDemand{
			Ask: "https://auth.example.com/check",
		},
	},
}
