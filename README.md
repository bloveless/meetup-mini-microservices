# meetup-mini-microservices
Using consul to create microservices which can discover each other using service discovery

To run this locally install consul and taskfile.

Run `task consul` to start the consule server in dev mode.

Run `task server` to start an http server.

Run `task server` to run another http server under the same service name as the first but with a different ID.

Finally run `task client` multiple times to make a request to the http server first by looking up the address from consule.
