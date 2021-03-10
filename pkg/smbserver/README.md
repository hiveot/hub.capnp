# Simple Message Bus Server 

The smbserver provides the internal simple message bus server of the WoST Hub. Intended for secure standalone pub/sub messaging between plugins.

This serves websockets over TLS. Messages have the following format:
  {command}:{channel}:{payload}

Where command is one of 'publish', 'receive', 'subscribe', 'unsubscribe'. See pkg/messaging/SmbusHelper.go for exported definitions.

This server only handles a single subscription per channel. The smbclient implements support for multiple subscribers on a channel.
