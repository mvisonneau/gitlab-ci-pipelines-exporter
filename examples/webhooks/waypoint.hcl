project = "gcpe-webhooks-example"

app "exporter" {
  build {
    use "docker" {}
  }

  deploy {
    use "docker" {
      service_port = 8080
    }
  }
}
