job "fabiolb-docs" {
  datacenters = [
    "nsvltn",
    "iplsin"
  ]
  type = "service"
  update {
    max_parallel = 1
    auto_revert = true
  }
  group "deploy" {
    count = 4
    constraint {
      attribute = "${node.datacenter}"
      operator = "distinct_property"
      value = "2"
    }
    restart {
      attempts = 3
      interval = "30m"
      delay = "15s"
      mode = "fail"
    }
    ephemeral_disk {
      size = 750
    }
    task "docker" {
      driver = "docker"
      config {
        image = "fabiolb/fabio-docs"
      }
      service {
        name = "fabiolb-docs"
        tags = [
          "urlprefix-fabiolb.net/", "urlprefix-www.fabiolb.net/"
        ]
        address_mode = "driver"
        port = 1180
        check {
          address_mode = "driver"
          port = 1180
          type = "http"
          path = "/check/ok"
          interval = "20s"
          timeout = "5s"
        }
      }
      env {
        SERVICE_IGNORE = "true"
      }
      resources {
        cpu = 500
        memory = 1024
      }
      logs {
        max_files = 5
        max_file_size = 100
      }
    }
  }
}
