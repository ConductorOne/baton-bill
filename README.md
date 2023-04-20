![Baton Logo](./docs/images/baton-logo.png)

# `baton-bill` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-bill.svg)](https://pkg.go.dev/github.com/conductorone/baton-bill) ![main ci](https://github.com/conductorone/baton-bill/actions/workflows/main.yaml/badge.svg)

`baton-bill` is a connector for Bill.com built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the Bill.com API to sync data about which users have access to various entities and organizations.

Check out [Baton](https://github.com/conductorone/baton) to learn more about the project in general.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-bill

baton-bill --username <username> --password <password> --organizationIds <firstOrgId> <secondOrgId> --developerKey <developerKey>
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_BILL_USERNAME=username -e BATON_BILL_PASSWORD=password -e BATON_BILL_ORGANIZATION_IDS=organizationIds -e BATON_BILL_DEVELOPER_KEY=developerKey ghcr.io/conductorone/baton-bill:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-bill/cmd/baton-bill@main

baton-bill --username <username> --password <password> --organizationIds <firstOrgId> <secondOrgId> --developerKey <developerKey>
baton resources
```

# Data Model

`baton-bill` will pull down information about the following Bill.com resources:

- Organizations
- Users
- Roles?

By default, `baton-bill` will sync information from any organizations that the provided credential has access to.

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-bill` Command Line Usage

```
baton-bill

Usage:
  baton-bill [flags]
  baton-bill [command]

Available Commands:
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
  -f, --file string             The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                    help for baton-bill
      --log-format string       The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string        The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --username string         The username to use to authenticate with Bill.com ($BATON_BILL_USERNAME)
      --password string         The password to use to authenticate with Bill.com ($BATON_BILL_PASSWORD)
      --organizationIds strings The organizationId to use to authenticate with Bill.com ($BATON_BILL_ORGANIZATION_IDS)
      --developerKey string     The developerKey to use to authenticate with Bill.com ($BATON_BILL_DEVELOPER_KEY)
  -v, --version                 version for baton-bill

Use "baton-bill [command] --help" for more information about a command.
```
