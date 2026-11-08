---
page_title: "caddy_tls Resource - terraform-provider-caddy"
subcategory: ""
description: |-
  Manages the Caddy TLS app configuration including certificate automation, certificate loading, and issuer configuration.
---

# Resource `caddy_tls`

Resource to manage the [Caddy TLS app](https://caddyserver.com/docs/json/apps/tls/). This resource manages TLS certificate automation, certificate loading from files or PEM strings, and issuer configuration (ACME, internal, ZeroSSL).

## Example Usage

### Automatic ACME certificates

```hcl
resource "caddy_tls" "tls" {
  certificates {
    automate = ["example.com", "www.example.com"]
  }

  automation {
    policy {
      subjects = ["example.com", "www.example.com"]

      issuer {
        acme {
          ca    = "https://acme-v02.api.letsencrypt.org/directory"
          email = "admin@example.com"
        }
      }
    }
  }
}
```

### Load certificates from files

```hcl
resource "caddy_tls" "tls" {
  certificates {
    load_files {
      certificate = "/etc/ssl/certs/example.pem"
      key         = "/etc/ssl/private/example.key"
      format      = "pem"
    }
  }
}
```

### ACME with Cloudflare DNS challenge

This uses the DNS-01 challenge via Cloudflare's API, which is useful when ports 80/443 are not reachable or for wildcard certificates. Requires building Caddy with the [caddy-dns/cloudflare](https://github.com/caddy-dns/cloudflare) module (e.g. via `xcaddy build --with github.com/caddy-dns/cloudflare`).

```hcl
variable "cloudflare_api_token" {
  type      = string
  sensitive = true
}

resource "caddy_tls" "tls" {
  certificates {
    automate = ["example.com", "*.example.com"]
  }

  automation {
    policy {
      subjects = ["example.com", "*.example.com"]

      issuer {
        acme {
          ca    = "https://acme-v02.api.letsencrypt.org/directory"
          email = "admin@example.com"

          challenges {
            dns {
              provider = jsonencode({
                name      = "cloudflare"
                api_token = var.cloudflare_api_token
              })
              resolvers = ["1.1.1.1", "8.8.8.8"]
            }
          }
        }
      }
    }
  }
}
```

> **Note:** The `provider` field accepts a JSON object where `name` identifies the DNS provider module and remaining fields are provider-specific. For Cloudflare, the fields are `api_token` (required, a scoped API token with `Zone.DNS:Edit` permission) and optionally `zone_token` (a token with `Zone:Read` permission).

### Internal (self-signed) certificates

```hcl
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
```

## Schema

### Optional

- **certificates** (Block Set, Max: 1) Certificate sources configuration.
  - **automate** (List of String) Domain names for which Caddy should automatically obtain and manage certificates.
  - **load_files** (Block List) Load certificate/key pairs from files on disk.
    - **certificate** (String, Required) Path to the certificate PEM file.
    - **key** (String, Required) Path to the private key file.
    - **format** (String) Certificate format (e.g., `pem`).
    - **tags** (List of String) Tags for the certificate.
  - **load_pem** (Block List) Load certificate/key pairs from PEM-encoded strings.
    - **certificate** (String, Required) PEM-encoded certificate.
    - **key** (String, Required) PEM-encoded private key.
    - **tags** (List of String) Tags for the certificate.

- **automation** (Block Set, Max: 1) Automatic certificate management configuration.
  - **policy** (Block List) Automation policies. Each policy defines how certificates for a set of subjects should be managed.
    - **subjects** (List of String) Domain names this policy applies to. If empty, applies as default.
    - **key_type** (String) Key type for certificates (e.g., `ec256`, `ec384`, `rsa2048`, `rsa4096`).
    - **must_staple** (Boolean) Whether to request OCSP Must-Staple.
    - **on_demand** (Boolean) Whether to enable on-demand TLS for this policy.
    - **issuer** (Block List) Certificate issuers for this policy. Each block must contain exactly one of `acme`, `internal`, or `zerossl`.
      - **acme** (Block Set, Max: 1) ACME issuer (Let's Encrypt, etc.).
        - **ca** (String) ACME CA directory URL.
        - **email** (String) Contact email for the ACME account.
        - **external_account** (Block Set, Max: 1) External Account Binding (EAB) credentials.
          - **key_id** (String, Required) EAB key ID.
          - **mac_key** (String, Required) EAB MAC key.
        - **challenges** (Block Set, Max: 1) ACME challenge configuration.
          - **http** (Block Set, Max: 1) HTTP-01 challenge settings.
            - **disabled** (Boolean) Disable HTTP-01 challenge.
            - **alternate_port** (Number) Alternate port for the challenge.
          - **tls_alpn** (Block Set, Max: 1) TLS-ALPN-01 challenge settings.
            - **disabled** (Boolean) Disable TLS-ALPN-01 challenge.
            - **alternate_port** (Number) Alternate port for the challenge.
          - **dns** (Block Set, Max: 1) DNS-01 challenge settings. Requires a Caddy build with a DNS provider module (e.g. `caddy-dns/cloudflare`).
            - **provider** (String, Required) JSON-encoded DNS provider config. Must include a `name` field identifying the provider module (e.g. `"cloudflare"`) plus provider-specific fields (e.g. `api_token`). Use `jsonencode()` in Terraform.
            - **ttl** (String) TTL for DNS records (Go duration string).
            - **propagation_delay** (String) Time to wait before starting propagation checks.
            - **propagation_timeout** (String) Maximum time to wait for DNS propagation.
            - **resolvers** (List of String) DNS servers to use for checking propagation (e.g. `["1.1.1.1"]`).
            - **override_domain** (String) Override the domain for the DNS challenge.
        - **trusted_roots_pem_files** (List of String) Paths to PEM files containing trusted root certificates.
      - **internal** (Block Set, Max: 1) Internal (self-signed) certificate issuer.
        - **ca** (String) CA ID (default: `local`).
        - **lifetime** (String) Certificate lifetime as a Go duration string (e.g., `8760h` for 1 year).
        - **sign_with_root** (Boolean) Whether to sign directly with the root certificate.
      - **zerossl** (Block Set, Max: 1) ZeroSSL-specific issuer.
        - **ca** (String) ZeroSSL CA directory URL.
        - **email** (String) Contact email.
        - **api_key** (String) ZeroSSL API key.
        - **external_account** (Block Set, Max: 1) EAB credentials.
          - **key_id** (String, Required) EAB key ID.
          - **mac_key** (String, Required) EAB MAC key.
  - **on_demand** (Block Set, Max: 1) Global on-demand TLS configuration.
    - **ask** (String) URL to check before obtaining a certificate on-demand.
    - **rate_limit** (Block Set, Max: 1) Rate limiting for on-demand certificate issuance.
      - **interval** (String) Time window for rate limiting (Go duration string).
      - **burst** (Number) Maximum number of certificates to issue within the interval.
