# Revere

[![codecov](https://codecov.io/gh/broadinstitute/revere/branch/main/graph/badge.svg?token=RLKHCZDWat)](https://codecov.io/gh/broadinstitute/revere)

Revere is a service communicating Terra's uptime and status to end-users
(like a metaphorical [Paul Revere](https://en.wikipedia.org/wiki/Paul_Revere%27s_Midnight_Ride)).

Revere cares about **services** and **components**. A **service** is an internal application or codebase, something that we can directly monitor. A **component** is a piece of user-facing functionality, something that impacts customers and clients.

Revere operates like this:
1. Accept status information about **services** from event sources:
   1. *[WIP: in-development]* Cloud Monitoring Alerts via Cloud Pub/Sub
2. Translate those events to impacts on **components**
3. Communicate those impacts to end-users:
   1. *[WIP: in-development]* Statuspage.io
    
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

Do not manually tag commits. This repository has [Bumper](https://github.com/DataBiosphere/github-actions/tree/master/actions/bumper) enabled, Git tags will be created automatically on merge to `main`. 
Patch version will be automatically bumped upon merge; including "`#minor`" or "`#major`" in the merge commit body will bump that instead.
