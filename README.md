# DLC Oracle

This project can serve as an oracle while forming Discreet Log Contracts. This oracle currently publishes the value of the US Dollar denominated in Bitcoin's smallest fraction (satoshis). You can interact with the oracle via simple REST calls. A live version of this oracle is running on [https://oracle.gertjaap.org/] 

If you want to learn more about Discreet Log Contracts, checkout the [whitepaper](https://adiabat.github.io/dlc.pdf)

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

You need to have golang installed, or you can use Docker

### Installing

First, clone the repository and install the dependencies

```
git clone https://github.com/gertjaap/dlcoracle
cd dlcoracle
go get -v ./...
```

Then you can build the oracle using
```
go build
```

### Running the oracle

Simply start the executable. Since the oracle generates a private key it will ask you for a password to protect it, that you have to enter each time you start up the oracle.

```
./dlcoracle
```

## REST Endpoints

| resource          | description                              |
|:------------------|:-----------------------------------------|
|[`/api/pubkey`](https://oracle.gertjaap.org/api/pubkey)      | Returns the public keys of the oracle     |
|[`/api/datasources`](https://oracle.gertjaap.org/api/datasources) | Returns an array of data sources the oracle publishes |
|[`/api/publication/{R}`](https://oracle.gertjaap.org/api/publication/1/1523447385) | Returns the value and signature published for data source point **R** (if published). R is hex encoded [33]byte |

## Using the public deployment

You're free to use my public deployment of the oracle as well. I have linked the URLs of the public deployment in the REST endpoint table above.

## Determine point R

Point R is determined by using the public keys of the oracle. R is the public key to the one-time-signing key of the message. It can be used to pre-compute the public key to any signed message. It is determined using:

```R = Q - h(t, Q)B```

Where Q and B can be found in the `/api/pubkey` response, and t is an encoding of the message type and timestamp. This is encoded by concatenating the datasourceid (uint64) and the unix timestamp of the publication time (uint64) together. Check [crypto/derivesign.go](crypto/derivesign.go) on how to implement this.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
