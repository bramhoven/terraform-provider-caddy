# terraform-provider-caddy

Terraform provider for managing the [Caddy Admin API](https://caddyserver.com/docs/api).

This project is maintained from the original
[conradludgate/terraform-provider-caddy](https://github.com/conradludgate/terraform-provider-caddy/)
with continued fixes and releases.

## Terraform Provider Source

Use this provider source in Terraform:

```hcl
terraform {
  required_providers {
    caddy = {
      source = "bramhoven/caddy"
    }
  }
}
```

## Setup

Run Caddy with the admin API enabled, then point the provider at that endpoint.

### Option 1: HTTP endpoint

The default Caddy admin endpoint is `http://localhost:2019`, which is also the provider default.
This is simple for local testing, but usually not recommended for remote/untrusted environments.

```hcl
provider "caddy" {
  host = "http://localhost:2019"
}
```

### Option 2: Unix socket (recommended)

Configure Caddy's admin endpoint to a unix socket, for example:
`unix//path/to/admin.sock`

Validate socket access:

```sh
curl -H "Host: " --unix-socket /path/to/admin.sock http://localhost/config/
```

Then configure the provider:

```hcl
provider "caddy" {
  host = "unix:///path/to/admin.sock"
}
```

### Optional: SSH tunneling

You can proxy the API connection over SSH when Caddy is on a remote host.

```hcl
provider "caddy" {
  host = "unix:///path/to/admin.sock"
  ssh {
    host     = "user@example.com:22" # port is required
    key_file = "~/.ssh/id_rsa"       # or use user:pass@example.com:22
    host_key = "example.com ssh_rsa AAAA..." # known_hosts format
  }
}
```

## Resources

### `caddy_http`

Manages the Caddy HTTP app configuration (ports, grace period).

### `caddy_server`

Manages individual HTTP servers with routes, error handlers, and logging.

### `caddy_tls`

Manages the Caddy TLS app configuration including certificate automation, certificate loading, and issuer configuration.

#### TLS with automatic ACME certificates

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

#### TLS with custom certificate files

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

#### TLS with Cloudflare DNS challenge (wildcard certs)

Requires building Caddy with [`caddy-dns/cloudflare`](https://github.com/caddy-dns/cloudflare).

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
          email = "admin@example.com"

          challenges {
            dns {
              provider = jsonencode({
                name      = "cloudflare"
                api_token = var.cloudflare_api_token
              })
            }
          }
        }
      }
    }
  }
}
```

#### TLS with internal (self-signed) certificates

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

#### TLS with ACME challenges and on-demand

```hcl
resource "caddy_tls" "tls" {
  automation {
    policy {
      subjects = ["example.com"]
      key_type = "ec256"

      issuer {
        acme {
          ca    = "https://acme-v02.api.letsencrypt.org/directory"
          email = "admin@example.com"

          external_account {
            key_id  = "your-kid"
            mac_key = "your-mac-key"
          }

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

    on_demand {
      ask = "https://auth.example.com/check"
      rate_limit {
        interval = "1m"
        burst    = 10
      }
    }
  }
}
```

#### Supported issuer types

- **`acme`** — Let's Encrypt, ZeroSSL, or any ACME-compatible CA. Supports `ca`, `email`, `external_account`, `challenges` (http, tls_alpn), and `trusted_roots_pem_files`.
- **`internal`** — Self-signed certificates from Caddy's built-in CA. Supports `ca`, `lifetime`, and `sign_with_root`.
- **`zerossl`** — ZeroSSL-specific issuer. Supports `ca`, `email`, `api_key`, and `external_account`.

### `caddy_server_route` (data source)

Outputs a route as JSON for use with `caddy_server`'s `routes` attribute.

## Example

See `example/main.tf` for a complete example using `caddy_server` and `caddy_server_route`.
