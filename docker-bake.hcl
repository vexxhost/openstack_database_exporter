variable "TAG" {
  default = "latest"
}

target "default" {
  dockerfile = "Dockerfile"

  tags = [
    "ghcr.io/vexxhost/openstack-database-exporter:latest",
    "ghcr.io/vexxhost/openstack-database-exporter:${TAG}",
  ]
}
