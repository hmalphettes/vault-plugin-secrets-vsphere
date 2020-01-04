module github.com/hmalphettes/vault-plugin-secrets-vsphere

go 1.13

replace github.com/nats-io/go-nats => github.com/nats-io/nats.go v1.9.1

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2

replace google.golang.org/cloud => cloud.google.com/go v0.39.0

require (
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/go-test/deep v1.0.2-0.20181118220953-042da051cf31
	github.com/google/uuid v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/go-hclog v0.10.1
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-retryablehttp v0.6.4 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2-0.20191001231223-f32f5fe8d6a8
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/vault/api v1.0.5-0.20191216174727-9d51b36f3ae4
	github.com/hashicorp/vault/sdk v0.1.14-0.20191218020134-06959d23b502
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/pierrec/lz4 v2.4.0+incompatible // indirect
	github.com/vmware/govmomi v0.21.0
	golang.org/x/crypto v0.0.0-20191219195013-becbf705a915 // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
	golang.org/x/sys v0.0.0-20191219235734-af0d71d358ab // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/genproto v0.0.0-20191216205247-b31c10ee225f // indirect
	google.golang.org/grpc v1.26.0 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
)
