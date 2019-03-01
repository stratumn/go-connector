# GO-CONNECTOR

The connector is a modular HTTP server acting as a middleware between the Trace API and a client application.

It is meant to be deployed on-premise on the client infrastructure, in order to benefit from the ability to access the decrypted data. It’s purpose is to handle all the logic that cannot be handled in Trace backend:

- Decryption
- Validation that cannot be performed sever side
- Replication
- Search
- Analytics

It leverages go plugins to offer flexibility in the way these modules are used.

## Build

Build the server binary
