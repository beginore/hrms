# Infrastructure
This directory is needed for interacting with additional pkgs & technologies that don't touch business logic, but help it to work, such as:

* App orchestration
* Config orchestration
* Errors setup (error struct, internal error codes)
* Storages setup, for example different DBs, their connection, migration management, storing migration files, public interface and mocks to interact with chosen DB for repositories (e.g. for different libs, such as sql, slqx, pgx)
* Observability setup (logger, tracing, metrics)
* Transport setup & errors mapping (gRPC, HTTP)