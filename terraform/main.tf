terraform {
  required_providers {
    graphapi = {
      version = "0.1.0"
      source  = "github.com/LeMikaelF/meetingsascode"
    }
  }
}

provider "graphapi" {
}

resource "graphapi_calendar_event" "example" {
  subject    = "Quick meeting to discuss the meaning of life"
  start_time = "2023-04-25T13:00:00"
  end_time   = "2023-04-25T13:30:00"
  time_zone  = "America/New_York"
  location   = "Dune"

  attendee {
    name    = "Mikaël Francoeur"
    address = "mikael@mikaelfrancoeur.com"
    type    = "required"
  }

  attendee {
    name    = "Paul Atreides"
    address = "kwisatz@arrakis.com"
    type    = "optional"
  }
}
