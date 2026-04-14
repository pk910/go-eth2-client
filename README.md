# go-eth2-client

[![Tag](https://img.shields.io/github/tag/ethpandaops/go-eth2-client.svg)](https://github.com/ethpandaops/go-eth2-client/releases/)
[![License](https://img.shields.io/github/license/ethpandaops/go-eth2-client.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/ethpandaops/go-eth2-client?status.svg)](https://godoc.org/github.com/ethpandaops/go-eth2-client)
![Lint](https://github.com/ethpandaops/go-eth2-client/workflows/golangci-lint/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethpandaops/go-eth2-client)](https://goreportcard.com/report/github.com/ethpandaops/go-eth2-client)

Go library providing an abstraction to multiple Ethereum beacon nodes. Its external API follows the official [Ethereum beacon APIs](https://github.com/ethereum/beacon-APIs) specification.

> **This is a fork.** `ethpandaops/go-eth2-client` is an ethPandaOps-maintained fork of [`attestantio/go-eth2-client`](https://github.com/attestantio/go-eth2-client). It preserves upstream's API so consumers can move between the two with minimal churn, while adding early support for Ethereum consensus forks that are still being specified.

## About this fork

The upstream project, maintained by Attestant Limited, is the canonical Go client for the beacon APIs and targets stable, spec-frozen forks. This fork exists alongside it for a narrower purpose: to let ethPandaOps tooling speak to consensus clients that implement in-development forks before those forks are frozen.

### Scope

- **Early fork support.** Types, SSZ encodings, and additional endpoints for upcoming forks (e.g. Fulu, Gloas, and subsequent in-development hard forks) are added here as soon as they are needed by ethPandaOps tooling running devnets and short-lived testnets.
- **Upstream compatibility, best-effort.** We track upstream changes and try to keep the public API source-compatible so downstream projects can swap between the two. This is a best-effort commitment, not a guarantee — when upstream and an in-development spec conflict, the spec wins for as long as it is WIP.
- **Not a replacement.** We do not intend to replace the upstream library. If you do not need pre-release fork support, prefer [`attestantio/go-eth2-client`](https://github.com/attestantio/go-eth2-client).

### Stability expectations

Because this fork follows specifications that are still moving:

- Types, field names, and endpoints related to in-development forks **will change** as the specs evolve — sometimes in breaking ways, including between patch releases.
- Stable mainnet fork types are only changed when upstream changes them or when an upstream merge forces it.
- Tagged releases are cut when ethPandaOps tooling needs a new baseline; there is no fixed cadence.

If you pin this library, pin it to an exact version and expect to revisit that pin whenever you upgrade a consensus client running an unreleased fork.

### Intended audience

This fork is built primarily for the [ethPandaOps](https://github.com/ethpandaops) tool suite, which sits close to Ethereum R&D and therefore needs a client library that can keep up with bleeding-edge consensus client builds. If you are running ethPandaOps tooling against devnets, this is the version you want. Otherwise, the upstream library is almost certainly the better choice.

## Install

`go-eth2-client` is a standard Go module which can be installed with:

```sh
go get github.com/ethpandaops/go-eth2-client
```

## Support

`go-eth2-client` supports beacon nodes that comply with the standard beacon node API. For in-development forks, the usable surface depends on what the client under test has implemented — a devnet-only fork is only exposed on clients that already speak it.

Tested against:

- [Lighthouse](https://github.com/sigp/lighthouse/)
- [Nimbus](https://github.com/status-im/nimbus-eth2)
- [Prysm](https://github.com/prysmaticlabs/prysm)
- [Teku](https://github.com/consensys/teku)
- [Grandine](https://github.com/grandinetech/grandine)

## Usage

Please read the [Go documentation for this library](https://godoc.org/github.com/ethpandaops/go-eth2-client) for interface information.

## Example

Below is a complete annotated example to access a beacon node.

```go
package main

import (
    "context"
    "errors"
    "fmt"

    eth2client "github.com/ethpandaops/go-eth2-client"
    "github.com/ethpandaops/go-eth2-client/api"
    "github.com/ethpandaops/go-eth2-client/http"
    "github.com/rs/zerolog"
)

func main() {
    // Provide a cancellable context to the creation function.
    ctx, cancel := context.WithCancel(context.Background())
    client, err := http.New(ctx,
        // WithAddress supplies the address of the beacon node, as a URL.
        http.WithAddress("http://localhost:5052/"),
        // LogLevel supplies the level of logging to carry out.
        http.WithLogLevel(zerolog.WarnLevel),
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("Connected to %s\n", client.Name())

    // Client functions have their own interfaces. Not all functions are
    // supported by all clients, so checks should be made for each function when
    // casting the service to the relevant interface.
    if provider, isProvider := client.(eth2client.GenesisProvider); isProvider {
        genesisResponse, err := provider.Genesis(ctx, &api.GenesisOpts{})
        if err != nil {
            // Errors may be API errors, in which case they will have more detail
            // about the failure.
            var apiErr *api.Error
            if errors.As(err, &apiErr) {
                switch apiErr.StatusCode {
                case 404:
                    panic("genesis not found")
                case 503:
                    panic("node is syncing")
                }
            }
            panic(err)
        }
        fmt.Printf("Genesis time is %v\n", genesisResponse.Data.GenesisTime)
    }

    // You can also access the struct directly if required.
    httpClient := client.(*http.Service)
    genesisResponse, err := httpClient.Genesis(ctx, &api.GenesisOpts{})
    if err != nil {
        panic(err)
    }
    fmt.Printf("Genesis validators root is %s\n", genesisResponse.Data.GenesisValidatorsRoot)

    // Cancelling the context passed to New() frees up resources held by the
    // client, closes connections, clears handlers, etc.
    cancel()
}
```

## Maintainers

This fork is maintained by the [ethPandaOps](https://github.com/ethpandaops) team, primarily [@pk910](https://github.com/pk910).

The upstream library was created and is maintained by Chris Berry ([@bez625](https://github.com/Bez625)) and contributors at Attestant Limited.

## Contribute

Contributions that target in-development fork support or keep this fork in sync with upstream are welcome. Please check [the issues](https://github.com/ethpandaops/go-eth2-client/issues). Bugs or improvements that are not fork-specific are usually a better fit for [upstream](https://github.com/attestantio/go-eth2-client) — we pull those changes in on a best-effort basis.

## License

[Apache-2.0](LICENSE)

- Original work © 2020–2025 Attestant Limited.
- Fork modifications © 2025 ethPandaOps.
