// Package privatebtc handles the creation of a Bitcoin private network using Docker.
//
// The package defines a NodeService interface that can be implemented to use different
// Node Handlers.
// The package also defines a RPCClientFactory interface that can be implemented to use
// different Go Bitcoin RPC Clients.
// A chain reorganisation manager is implemented, it streamlines the process of creating
// chain reorganisations.
package privatebtc
