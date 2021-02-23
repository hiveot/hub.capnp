# Simple Message Bus server plugin

This smbserver plugin provides the internal simple message bus server the WoST gateway. Intended for secure standalone pub/sub messaging between plugins.

Usage: 
This server can be started via the gateway or directly from the commandline. Run with --help to view the available commandline options.
> dist/bin/smbserver --help



This serves websockets over TLS. Messages have the following format:
  {command}:{channel}:{payload}

Where command is one of 'publish', 'receive', 'subscribe', 'unsubscribe'. See pkg/messaging/SmbusHelper.go for exportedd definitions.

This server only handles a single subscription per channel. The smbus client implements support for multiple subscribers on a channel.
