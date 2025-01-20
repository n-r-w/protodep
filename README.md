# protodep - Protocol Buffers Dependency Management Tool

A powerful dependency management tool for Protocol Buffers IDL files (.proto), forked from [stormcat24/protodep](https://github.com/stormcat24/protodep) with significant enhancements.

## Project Overview

protodep helps you manage Protocol Buffer dependencies in microservice architectures by:

- Vendoring .proto files from Git repositories
- Supporting multiple version control strategies (branch, revision)
- Managing complex dependency structures with flexible configuration
- Supporting both local and remote proto files

## Key Features

- **Multiple Authentication Methods**:
  - Git credential helpers support
  - .netrc file authentication
  - SSH key authentication  
  - HTTPS with basic auth
  - Environment variable-based authentication
- **Flexible Dependency Management**:
  - Local proto files import
  - GitLab subgroups support
  - Selective file inclusion/exclusion
  - Branch or revision-based versioning
  - Protocol-specific configuration

## Installation

```bash
go install -v github.com/n-r-w/protodep@latest
```

## Usage

### Configuration (protodep.toml)

Create a `protodep.toml` file in your project root:

```toml
# Base output directory for vendored proto files
proto_outdir = "./proto"

# Remote repository dependency
[[dependencies]]
  target = "github.com/org/repo/protos" # Remote repository
  branch = "master"             # Use branch or revision
  path = "path/to/protos"       # Target local subdirectory containing proto files
  ignores = ["./ignored-dir"]   # Optional: Directories to ignore
  includes = ["some.proto"]     # Optional: Files to include
  protocol = "ssh"              # Optional: Protocol to use (ssh/https)

# GitLab with subgroups
[[dependencies]]
  target = "gitlab.company.org/group/subgroup/repo/protos"
  subgroup = "subgroup"
  path = "path/to/protos"
  revision = "v1.0.0"

# Local directory dependency
[[dependencies]]
  local_folder = "./api/broker"
  path = "proto/broker"

# Environment variable authentication
[[dependencies]]
  target = "github.com/org/private-repo"
  path = "protos"
  username_env = "GITHUB_USERNAME"  # Username from environment variable
  password_env = "GITHUB_TOKEN"     # Token from environment variable  
```

### Authentication Options

1. **HTTPS with Basic Auth**:

```bash
protodep up --use-https \
  --basic-auth-username=your-username \
  --basic-auth-password=your-token
```

2. **SSH Key Authentication**:

```bash
protodep up -i ~/.ssh/id_rsa -p "ssh-key-password"
```

3. **.netrc File** (Create in home directory):

```plaintext
machine github.com
login your-username
password your-token
```

4. **Environment Variables** (In protodep.toml):

```toml
[[dependencies]]
  target = "github.com/org/repo"
  username_env = "GITHUB_USERNAME"    # Environment variable for username
  password_env = "GITHUB_TOKEN"       # Environment variable for token/password  
```

### Command Line Options

```bash
protodep up [flags]

Flags:
  -i, --identity-file string      SSH identity file path
  -p, --password string           SSH key password
  -c, --cleanup                   Cleanup cache before execution
  -u, --use-https                 Use HTTPS instead of SSH
  -n, --use-netrc                 Use .netrc file for authentication (default: true)
  -m, --use-git-credentials      Use git credentials helper (default: true)
      --basic-auth-username      HTTPS basic auth username
      --basic-auth-password      HTTPS basic auth password/token
```

Note: Both `use-netrc` (-n) and `use-git-credentials` (-m) are enabled by default with `use-git-credentials` priority. Use the respective flags to disable them if needed.
