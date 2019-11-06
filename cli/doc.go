// Package cli is like the "main" package. It's the responsible of
// starting all the services, but it's not the responsible of initializing
// things, such as DB connections, or reading the configuration. Everything
// has to be working before calling cli.Main
package cli
