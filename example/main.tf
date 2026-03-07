terraform {
  required_providers {
    caddy = {
      version = "~> 0.2.0"
      source  = "bramhoven/caddy"
    }
  }
}

provider "caddy" {
  host            = "unix:///tmp/caddy-admin.sock"
  ssh {
    host = "terraform@ssh.example.com:22"
    key_file = "/home/oon/.ssh/terraform"
  }
}

resource "caddy_server" "foo" {
  name   = "foo"
  listen = [":443"]

  routes = [
    data.caddy_server_route.route1.id,
  ]
}

data "caddy_server_route" "route1" {
  match {
    host = ["example1.bramhoven.dev"]
  }

  handle {
    reverse_proxy {
      upstream {
        dial = "localhost:8080"
      }
    }
  }
}
