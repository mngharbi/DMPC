# Distributed Multiuser Private Channels [![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0) ![Travis CI Build Status](https://api.travis-ci.org/mngharbi/DMPC.svg?branch=master)

## Overview
DMPC enables end-to-end encrypted, authenticated, channel based communication. It can be thought of as an extension to TCP. Users are identified by their public keys fingerprints. Authentication is based on a web of trust approach. Channels have multiple users as opposed to bi-directional sockets. It's aware of user permissions both at the global level *(e.g. whether you trust a user to introduce other users to you)* and at the channel level *(e.g. whether a participating user can close the channel)*.

It's built to be general enough to accommodate any type of setup but decouples communication from the underlying network. It's also built with distributed systems in mind, as state is eventually consistent regardless of the order operations come in.

## Encryption
All messages are wrapped into operations that have two layers of encryption.

The outer layer, also called transaction, is meant for data in transit. It uses a temporary symmetric key (encrypted using the recipient's public key).

The inner layer, also called operation, is meant for data at rest. Every channel has its own symmetric encryption key (only the participants have it), and it's used for all communication within the channel.

## Installation

```
go get -t github.com/mngharbi/DMPC
dmpc install
```

## Usage

To start the ingestion pipeline *(a websocket)*
```
dmpc server
```

## Dependencies

Apart from the packages implemented in this repo, DMPC depends on [mngharbi/memstore](https://github.com/mngharbi/memstore), [mngharbi/gofarm](https://github.com/mngharbi/gofarm), [gorilla/websocket](https://github.com/gorilla/websocket), [rs/xid](https://github.com/rs/xid), and [urfave/cli](https://github.com/urfave/cli).
