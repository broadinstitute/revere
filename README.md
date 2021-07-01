# Revere

Revere is a service communicating Terra's uptime and status to end-users
(like a metaphorical [Paul Revere](https://en.wikipedia.org/wiki/Paul_Revere%27s_Midnight_Ride)).

Revere accepts events as runtime input and translates those events to what communication channels expect:
- Supported event sources:
    - *[WIP: planned]* Automatic: Cloud Monitoring Alerts via Cloud Pub/Sub
    - *[WIP: planned]* Manual: authenticated REST API
- Supported communication platforms:
    - *[WIP: in-development]* Statuspage.io
    
## Usage

Revere is started via CLI:

```shell
go run main.go
```

Docker images are built automatically and are uploaded to [dsp-artifact-registry](https://console.cloud.google.com/artifacts/docker/dsp-artifact-registry/us-central1/revere).

### Configuration

[Viper](https://github.com/spf13/viper) reads the configuration file and allows a wide range of formats; documentation here uses YAML.

Viper reads a file given by `--config` via CLI, or looks for `revere.yaml` in the following directories:
1. `./`
2. `/etc/revere/`

The configuration file's format is defined by [`internal/config.go`](https://github.com/broadinstitute/revere/tree/main/internal/configuration/config.go).


## Development

### Repository Structure

```
.
├── cmd/
│   └── # CLI commands
├── internal/
│   └── # Service operation [stub, more to be added here]
└── main.go # CLI entrypoint
```

### Git Policy

`go fmt` should be run before pushed commits.

Do not manually tag commits. This repository has [Bumper](https://github.com/DataBiosphere/github-actions/tree/master/actions/bumper) enabled, Git tags will be created automatically on merge to `main`. See [Bumper](https://github.com/DataBiosphere/github-actions/tree/master/actions/bumper) for info on controlling this process.

Configurations for GoLand/VSCode may (tentatively) be committed, transient components are part of the
generated `.gitiginore`.