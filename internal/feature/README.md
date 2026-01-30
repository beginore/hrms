# FEATURES
This directory is needed for writing our business logic.
Here we will have 3 layered structure.
* Layer 1 (transport) is responsible for various protocols we use to communicate with the outside world.
  Transport layer works only with service methods and converts DTOs into Domains.
* Layer 2 (service) will be the largest one — the business logic layer, where we write absolutely all logic.
  Service can take different service methods and only its repository methods. Converts parts of Domain to needed Models.
* Layer 3 (repository) is the layer for working with various databases / storages.
  Here we work with Models that represent the table schema