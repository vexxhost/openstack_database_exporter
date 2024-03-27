FROM golang:1.22.1 AS builder
WORKDIR /src
COPY go.mod go.sum /src
RUN go mod download
COPY . /src
RUN CGO_ENABLED=0 go build -o /openstack_database_exporter

FROM scratch
COPY --from=builder /openstack_database_exporter /bin/openstack_database_exporter
EXPOSE 9180
ENTRYPOINT ["/bin/openstack_database_exporter"]
