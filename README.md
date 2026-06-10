# ngrok Terraform Provider (v2)

A Terraform provider for managing [ngrok](https://ngrok.com) resources, built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

This is a ground-up rewrite of the [original ngrok provider](https://github.com/ngrok/terraform-provider-ngrok), replacing the legacy SDKv2 with the Plugin Framework and the hand-rolled REST client with the official [ngrok-api-go/v9](https://github.com/ngrok/ngrok-api-go) client.

## Usage

```hcl
terraform {
  required_providers {
    ngrok = {
      source  = "ngrok/ngrok"
      version = "~> 1.0"
    }
  }
}

provider "ngrok" {}

resource "ngrok_reserved_domain" "example" {
  domain      = "app.example.com"
  description = "Production app domain"
}

resource "ngrok_cloud_endpoint" "example" {
  url = "https://${ngrok_reserved_domain.example.domain}"

  traffic_policy = jsonencode({
    on_http_request = [
      {
        actions = [
          {
            type   = "custom-response"
            config = { status_code = 200, content = "Hello from ngrok!" }
          }
        ]
      }
    ]
  })
}
```

Configure the API key via the `NGROK_API_KEY` environment variable or in the provider block.

## Resources

| Resource | Description |
|---|---|
| `ngrok_api_key` | API keys for authenticating to the ngrok API |
| `ngrok_agent_ingress` | Custom agent ingress domains |
| `ngrok_certificate_authority` | Certificate authorities for mTLS |
| `ngrok_cloud_endpoint` | Cloud endpoints with traffic policy |
| `ngrok_credential` | Tunnel authtokens |
| `ngrok_event_destination` | Event destination targets (Datadog, AWS, Azure) |
| `ngrok_event_subscription` | Event subscriptions |
| `ngrok_ip_policy` | IP policy groups |
| `ngrok_ip_policy_rule` | IP policy CIDR rules |
| `ngrok_ip_restriction` | IP restrictions on API/dashboard/agent/endpoints |
| `ngrok_kubernetes_operator` | Kubernetes operator registration |
| `ngrok_reserved_addr` | Reserved TCP addresses |
| `ngrok_reserved_domain` | Reserved domains |
| `ngrok_secret` | Secrets stored in vaults |
| `ngrok_service_user` | Service users (bot users) |
| `ngrok_ssh_certificate_authority` | SSH certificate authorities |
| `ngrok_ssh_credential` | SSH credentials |
| `ngrok_ssh_host_certificate` | SSH host certificates |
| `ngrok_ssh_user_certificate` | SSH user certificates |
| `ngrok_tls_certificate` | TLS certificates |
| `ngrok_vault` | Secret management vaults |

Every resource also has a corresponding **data source** for lookups.

## What changed from v0.x

- **Terraform Plugin Framework** replaces legacy SDKv2
- **Official `ngrok-api-go/v9` client** replaces hand-rolled REST client
- **Cloud endpoints** replace edges — all edge, backend, and endpoint configuration resources are removed
- **New resources**: `ngrok_cloud_endpoint`, `ngrok_kubernetes_operator`, `ngrok_vault`, `ngrok_secret`
- **Data sources** for every resource
- **Import support** for every resource

See the [upgrade guide](docs/guides/version-1-upgrade.md) for migration details.

## Development

```bash
# Build and install locally
make install

# Run tests
go test ./...

# Run acceptance tests (requires NGROK_API_KEY)
TF_ACC=1 NGROK_API_KEY=your-key go test ./... -v -timeout 120m
```

After `make install`, configure your test HCL to use the local provider:

```hcl
terraform {
  required_providers {
    ngrok = {
      source  = "ngrok/ngrok"
      version = "0.0.1"
    }
  }
}
```

Then run `rm -f .terraform.lock.hcl && terraform init` to pick up the local build.
