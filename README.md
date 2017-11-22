# Cloud Foundry Management (cf-mgmt)

Go automation for managing orgs, spaces, users (from ldap groups or internal store) mapping to roles, quotas, application security groups and private-domains that can be driven from concourse pipeline and GIT managed metadata

## Maintainer

* [Caleb Washburn](https://github.com/calebwashburn)

## Support

cf-mgmt is a community supported cloud foundry add-on.  Opening issues for questions, feature requests and/or bugs is the best path to getting "support".  We strive to be active in keeping this tool working and meeting your needs in a timely fashion.

## Install

Compiled releases are available on Github.
Download the binary for your platform and place it somewhere on your path.
Don't forget to `chmod +x` the file on Linux and macOS.

Alternatively, you may wish to build from source.

## Build from the source

`cf-mgmt` is written in [Go](https://golang.org/).
To build the binary yourself, follow these steps:

* Install `Go`.
* Install [Glide](https://github.com/Masterminds/glide), a dependency management tool for Go.
* Clone the repo:
  - `mkdir -p $(go env GOPATH)/src/github.com/pivotalservices`
  - `cd $(go env GOPATH)/src/github.com/pivotalservices`
  - `git clone git@github.com:pivotalservices/cf-mgmt.git`
* Install dependencies:
  - `cd cf-mgmt`
  - `glide install`
  - `go build -o cf-mgmt cmd/cf-mgmt/main.go`
  - `go build -o cf-mgmt-config cmd/cf-mgmt-config/main.go`

To cross compile, set the `$GOOS` and `$GOARCH` environment variables.
For example: `GOOS=linux GOARCH=amd64 go build`.

## Testing

To run the unit tests, use `go test $(glide nv)`.

### Integration tests

There are integration tests that require some additional configuration.

The LDAP tests require an LDAP server, which can be started with Docker:

```
docker pull cwashburn/ldap
docker run -d -p 389:389 --name ldap -t cwashburn/ldap
RUN_INTEGRATION_TESTS=true go test ./ldap/...
```

The remaining integration tests require [PCF Dev](https://pivotal.io/pcf-dev)
to be running and the CF CLI to be installed.

```
cf dev start
uaac target uaa.local.pcfdev.io
uaac token client get admin -s admin-client-secret
uaac client add cf-mgmt \
  --name cf-mgmt \
  --secret cf-mgmt-secret \
  --authorized_grant_types client_credentials,refresh_token \
  --authorities cloud_controller.admin,scim.read,scim.write
RUN_INTEGRATION_TESTS=true go test ./integration/...
```

## Code Generation

Some portions of this code are autogenerated.  To regenerate them, install the prerequisites:

- `go get -u github.com/jteeuwen/go-bindata`
- `go get -u github.com/golang/mock/mockgen`
- `go get -u github.com/maxbrunsfeld/counterfeiter`

And then run `go generate $(glide nv)` from the project directory, or `go generate .`
from a specific directory.

## Contributing

PRs are always welcome or open issues if you are experiencing an issue and will do my best to address issues in timely fashion.

## Documentation

- See [here](docs/README.md) for documentation on all the available commands for running cf-mgmt
- See [here](docs/config/README.md) for documentation on all the configuration documentation and commands
