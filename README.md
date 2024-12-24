protodep - dependency tool for Protocol Buffers IDL file (.proto) vendoring tool
=======

## Evolution of <https://github.com/n-r-w/protodep>, which is the evolution of <https://github.com/stormcat24/protodep>

## What's new in this fork (compared to n-r-w fork)
- [x] Added support for git credential helpers

### Stuff added in n-r-w's fork
- [x] Added support for .netrc file
- [x] Added support for local proto files import
- [x] Added support for gitlab subgroups
- [x] `protodep.lock` file removed
- [x] `-f`, `--force` option removed

## Motivation

In building Microservices architecture, gRPC with Protocol Buffers is effective. When using gRPC, your application will depend on many remote services.

If you manage proto files in a git repository, what will you do? Most remote services are managed by git and they will be versioned. We need to control which dependency service version that application uses.

## Install

### go install

```bash
go install -v github.com/torqio/protodep@latest
```

## Usage

### protodep.toml

Proto dependency management is defined in `protodep.toml`.

```Ruby
proto_outdir = "./proto"

[[dependencies]]
  target = "github.com/stormcat24/protodep/protobuf"
  branch = "master"

[[dependencies]]
  target = "github.com/grpc-ecosystem/grpc-gateway/examples/examplepb"
  revision = "v1.2.2"
  path = "grpc-gateway/examplepb"

# blacklist by "ignores" attribute
[[dependencies]]
  target = "github.com/kubernetes/helm/_proto/hapi"
  branch = "master"
  path = "helm/hapi"
  ignores = ["./release", "./rudder", "./services", "./version"]
  
# whitelist by "includes" attribute
[[dependencies]]
  target = "github.com/protodep/catalog/hierarchy"
  branch = "main"
  includes = [
    "/protodep/hierarchy/service.proto",
    "**/fuga/**",
  ]
  protocol = "https"

# gitlab subgroups
# repository: gitlab.company.org/core/product/backend/service1
# subgroup: backend/service
[[dependencies]]
  target = "gitlab.company.org/core/product/backend/service1/api/protos"
  subgroup = "backend/service1"
  revision = "v1.0.0"
  path = "service1"
  username_env = "GITLAB_USERNAME" # user environment variable for HTTP Basic Authentication
  password_env = "GITLAB_PASSWORD" # token/password environment variable for HTTP Basic Authentication

# import from local folder
[[dependencies]]
local_folder = "./api/broker"
path = "proto/broker"
```

### protodep up

In same directory, execute this command.

```bash
protodep up
```

### Getting via HTTPS

If you want to get it via HTTPS, do as follows.

```bash
protodep up --use-https
```

And also, if Basic authentication is required, do as follows.
If you have 2FA enabled, specify the Personal Access Token as the password.

```bash
$ protodep up --use-https \
    --basic-auth-username=your-github-username \
    --basic-auth-password=your-github-password
```

You can also set the username and password as environment variables and set them in the protodep.toml file for each dependency.

```toml
[[dependencies]]
  target = "github.com/your-org/your-repo"
  branch = "master"
  username_env = "GITHUB_USERNAME"
  password_env = "GITHUB_PASSWORD"
```

Another way is to use the .netrc file in your home directory. Set the username and password in the .netrc as follows.

```bash
machine github.com
login your-github-username
password your-github-token
```

### License

Apache License 2.0, see [LICENSE](https://github.com/stormcat24/protodep/blob/master/LICENSE).
