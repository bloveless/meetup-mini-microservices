# meetup-mini-microservices
Using consul to create microservices which can discover each other using service discovery

To run this locally install consul and taskfile.

Run `task consul` to start the consule server in dev mode.

Run `task server-one` to start an http server implementing the first algorithm.

Run `task server-two` to start an http server implementing the second algorithm.

Finally run `task instigator` to start sending requests and listening for the final ring on the telephone.
