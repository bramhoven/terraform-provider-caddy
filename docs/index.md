---
page_title: "caddy Provider"
subcategory: ""
description: |-
  
---

# caddy Provider

Terraform provider for managing the [Caddy Admin API](https://caddyserver.com/docs/api).

## Setup

Start Caddy with the admin API enabled, then configure the provider to connect to it.

Use this provider source:

```hcl
terraform {
  required_providers {
    caddy = {
      source = "bramhoven/caddy"
    }
  }
}
```

### HTTP Endpoint

The simplest option is the default endpoint `http://localhost:2019`.
This is useful for local development.

```hcl
provider "caddy" {
  host = "http://localhost:2019"
}
```

### Unix Sockets

The recommended option is a unix socket endpoint, for example:
`unix//path/to/admin.sock`.

Once Caddy is running, validate access:

```sh
curl -H "Host: " --unix-socket /path/to/admin.sock http://localhost/config/
```

Then configure the provider:

```hcl
provider "caddy" {
  host = "unix:///path/to/admin.sock"
}
```

### SSH

In addition to the host endpoint above, you can proxy API traffic through SSH.

```hcl
provider "caddy" {
  host = "unix:///path/to/admin.sock"
  ssh {
    host     = "user@example.com:22" # port is required
    key_file = "~/.ssh/id_rsa"       # or specify user:pass@example.com:22
    host_key = "example.com ssh_rsa AAAA..." # known_hosts format
  }
}
```

## Schema

### Optional

- **host** (String) http or unix socket to act as the caddy admin API endpoint
- **ssh** (Block Set) (see [below for nested schema](#nestedblock--ssh))

<a id="nestedblock--ssh"></a>
### Nested Schema for `ssh`

SSH Configuration to proxy caddy API access through. Recommended if accessing caddy API remotely as caddy provides no authentication

Required:

- **host** (String)
- **host_key** (String)
- **key_file** (String)
