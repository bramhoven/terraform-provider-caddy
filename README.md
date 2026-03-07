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

## Example

See `example/main.tf` for a complete example using `caddy_server` and `caddy_server_route`.
