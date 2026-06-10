terraform {
  required_providers {
    ngrok = {
      source  = "ngrok/ngrok"
      version = "0.0.1"
    }
  }
}

provider "ngrok" {}

# ─── Event Subscription ──────────────────────────────────────────────────────
resource "ngrok_event_subscription" "test" {
  description = "TF test event subscription updated"
  metadata    = jsonencode({ test = true })
  sources = [
    { type = "ip_policy_created.v0" },
    { type = "ip_policy_updated.v0" },
    { type = "ip_policy_deleted.v0" },
  ]
  destination_ids = [ngrok_event_destination.test.id]
}

output "event_subscription_id" {
  value = ngrok_event_subscription.test.id
}

# ─── Event Destination ───────────────────────────────────────────────────────
resource "ngrok_event_destination" "test" {
  description = "TF test event destination"
  metadata    = jsonencode({ test = true })
  format      = "json"
  target = {
    datadog = {
      api_key = "835860fe93b5ef34435c79e9527fca85"
      ddsite  = "US5"
      ddtags  = "env:test,team:platform"
      service = "ngrok-tf-test-updated"
    }
  }
}

output "event_destination_id" {
  value = ngrok_event_destination.test.id
}

# ─── Reserved Domain (no cert policy — tests null-vs-unknown fix) ─────────────
resource "ngrok_reserved_domain" "test" {
  domain      = "tf-test-manual.ngrok.app"
  description = "TF test reserved domain"
  metadata    = jsonencode({ test = true })
}

output "reserved_domain_id" {
  value = ngrok_reserved_domain.test.id
}


