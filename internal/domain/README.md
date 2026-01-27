# Domain

This directory is needed for writing domain structs.
They represent the business entities.<br>
For example, we write here User, and it will consist of needed aggregated fields from different user models.<br>
Shared across services to avoid cyclic imports.<br>
Additionally, business errors can be created in the same file where the domain is set.