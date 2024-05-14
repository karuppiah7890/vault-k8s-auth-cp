# vault-k8s-auth-cp

Using this CLI tool, you can copy Vault Kubernetes Auth Config and Roles from one Kubernetes Auth Method mount path to another mount path, even across Vault instances! :D

Note: The tool is written in Golang and uses Vault Official Golang API. The official Vault API documentation is here - https://pkg.go.dev/github.com/hashicorp/vault/api

Note: The tool needs Vault credentials of a user/account that has access to both the source and destination Vault, to read the Kubernetes Auth Method Configuration and Roles from source Vault and to configure the Kubernetes Auth Method Configuration and Roles in the destination Vault

Note: We have tested this only with some versions of Vault (like v1.15.x). So beware to test this in a testing environment before using this in critical environments like production!

Note ⚠️‼️🚨: If the destination Vault mount path has some configuration, it will be overwritten! This is true for both Kubernetes Auth Method Configuration and Roles. All the configuration and roles in source Vault Kubernetes Auth Method will be present in destination Vault. If the destination Vault has some extra Roles configured, it might have those. This scenario has NOT been tested currently

Note: This does NOT copy the Token Reviewer JWT token as it's not supported to be read from Vault once written. I think this is a security feature that Vault has implemented to not be able to read the token reviewer JWT token once it has been configured, to avoid leakage/exposure of the token

Note: This tool does NOT currently enable auth methods at any path. It assumes that the Kubernetes Auth Method has already been enabled at the given destination path and just configures it and adds roles

Note: This method also does NOT create any Vault policies or copy Vault policies from source Vault to detination Vault. It assumes that the Vault policies referred to in the Kubernetes Auth Method Roles (Roles Configuration) in the source Vault already exist in the destination Vault. If the policies referred to in the Kubernetes Auth Method Roles exist in source Vault but not in destination Vault, but you copy the Kubernetes Auth Method Configuration and Roles Configuration from source Vault to destination Vault, the Kubernetes Auth Method in the destination Vault will not work as expected, that is - it will not work the same way as the source Vault, since the destination Vault will not have the Vault policies referred to in the Kubernetes Auth Method Roles which was just copied from source Vault to destination Vault

Future version ideas:
- Support for copying Vault Policies referred to in the Kubernetes Auth Method Roles (Roles Configuration) from source Vault to destination Vault
- Support for providing the Token Reviewer JWT Token as user input, say through environment variable, so that you can copy that to destination Vault and configure it as part of Kubernetes Auth Method Configuration, since it CANNOT be read from the source Vault through the Vault API
- Remove log verbosity for default settings. It's too verbose now, by default.
- Remove log which shows no information (`nil`) about the data copied to destination Vault, as Write API does NOT seem to be returning any data in the reponse, so, we can just ignore it

## Building

```bash
go build -v
```

## Usage

```bash
$ ./vault-k8s-auth-cp
usage: vault-k8s-auth-cp <source-k8s-auth-mount-path> <destination-k8s-auth-mount-path>
```

# Demo

Source Vault, it's a secured Vault with HTTPS API enabled and a big token for root. It has some auth methods enabled and configured.

I'm using the Vault Root Token here for full access

```bash
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_TOKEN="some-big-token-here"
$ export VAULT_CACERT=$HOME/vault-ca.crt

$ vault status
Key                     Value
---                     -----
Seal Type               shamir
Initialized             true
Sealed                  false
Total Shares            5
Threshold               3
Version                 1.15.6
Build Date              2024-02-28T17:07:34Z
Storage Type            raft
Cluster Name            vault-cluster-9f170feb
Cluster ID              151e903e-e1e7-541e-d089-ce8db2da0a34
HA Enabled              true
HA Cluster              https://karuppiah-vault-0:8201
HA Mode                 active
Active Since            2024-04-27T23:15:36.130464099Z
Raft Committed Index    78945
Raft Applied Index      78945

$ vault auth list
Path           Type          Accessor                    Description                Version
----           ----          --------                    -----------                -------
kubernetes/    kubernetes    auth_kubernetes_03e3ab8d    kubernetes backend         n/a
production/    kubernetes    auth_kubernetes_0666b235    n/a                        n/a
token/         token         auth_token_b827a290         token based credentials    n/a

$ vault read auth/production/config
Key                       Value
---                       -----
disable_iss_validation    true
disable_local_ca_jwt      true
issuer                    n/a
kubernetes_ca_cert        -----BEGIN CERTIFICATE-----
something
-----END CERTIFICATE-----
kubernetes_host           https://some-kubernetes-cluster.com
pem_keys                  []

$ vault list auth/production/role
Keys
----
default

$ vault read auth/production/role/default
Key                                 Value
---                                 -----
alias_name_source                   serviceaccount_uid
bound_service_account_names         [default]
bound_service_account_namespaces    [default]
policies                            [allow_secrets]
token_bound_cidrs                   []
token_explicit_max_ttl              0s
token_max_ttl                       0s
token_no_default_policy             false
token_num_uses                      0
token_period                        0s
token_policies                      [allow_secrets]
token_ttl                           1h
token_type                          default
ttl                                 1h

$ vault policy list
allow_secrets
default
root

$ vault policy read allow_secrets
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

We can see that it has some auth methods enabled and configured. We can also see that it has some policies configured and a Kubernetes Auth Method Role configured for one of the Kubernetes Auth Methods. We are interested in the Kubernetes Auth Method enabled at `production` mount path of the source Vault.

Destination Vault, it's a local dev Vault server, with no HTTPS, with token as `root`. It was run using

```bash
$ vault server -dev -dev-root-token-id root -dev-listen-address 127.0.0.1:8300
```

We see that it has no auth enabled.

I'm using the Vault Root Token here for full access

```bash
$ export VAULT_ADDR='http://127.0.0.1:8300'
$ export VAULT_TOKEN="root"

$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    1
Threshold       1
Version         1.15.4
Build Date      2023-12-04T17:45:28Z
Storage Type    inmem
Cluster Name    vault-cluster-4543cfc5
Cluster ID      c521d162-70ba-bf83-c33d-c6283366d735
HA Enabled      false

$ vault auth list
Path      Type     Accessor               Description                Version
----      ----     --------               -----------                -------
token/    token    auth_token_517e4e2e    token based credentials    n/a
```

Let's enable Kubernetes Auth over here at `production-copy` mount path and see it's configuration

```bash

$ vault auth enable -path production-copy kubernetes
Success! Enabled kubernetes auth method at: production-copy/

$ vault auth list
Path                Type          Accessor                    Description                Version
----                ----          --------                    -----------                -------
production-copy/    kubernetes    auth_kubernetes_f09ee9e8    n/a                        n/a
token/              token         auth_token_517e4e2e         token based credentials    n/a

$ vault read auth/production-copy/config
No value found at auth/production-copy/config

$ vault list auth/production-copy/role
No value found at auth/production-copy/role
```

As we can see it has no configurations or roles defined/configured, since we just enabled it. Now, let's just copy the Kubernetes Auth Method Configuration and Roles from source Vault to destination Vault. Again, note that we are interested in the Kubernetes Auth Method enabled at `production` mount path of the source Vault.

I'm using the Vault Root Token here for both source Vault and destination Vault, for full access. But you don't need Vault Root Token. You just need any Vault Token / Credentials that has enough access to read Kubernetes Auth Method Configuration and Roles from source Vault and configure (write) Kubernetes Auth Method Configuration and Roles in destination Vault

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="some-big-token-here"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="root"

$ ./vault-k8s-auth-cp production production-copy

Source Kubernetes Auth Config: map[disable_iss_validation:true disable_local_ca_jwt:true issuer: kubernetes_ca_cert:-----BEGIN CERTIFICATE-----
something
-----END CERTIFICATE-----
 kubernetes_host:https://some-kubernetes-cluster.com pem_keys:[]]

Destination Kubernetes Auth Config: <nil>

Source Kubernetes Auth Roles: map[keys:[default]]

Source Kubernetes Auth Role: map[alias_name_source:serviceaccount_uid bound_service_account_names:[default] bound_service_account_namespaces:[default] policies:[allow_secrets] token_bound_cidrs:[] token_explicit_max_ttl:0 token_max_ttl:0 token_no_default_policy:false token_num_uses:0 token_period:0 token_policies:[allow_secrets] token_ttl:3600 token_type:default ttl:3600]

Destination Kubernetes Auth Role: <nil>
```

Now, let's look at the destination Vault and see if everything is copied and hence configured

```bash
$ export VAULT_ADDR='http://127.0.0.1:8300'
$ export VAULT_TOKEN="root"

$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    1
Threshold       1
Version         1.15.4
Build Date      2023-12-04T17:45:28Z
Storage Type    inmem
Cluster Name    vault-cluster-4543cfc5
Cluster ID      c521d162-70ba-bf83-c33d-c6283366d735
HA Enabled      false

$ vault auth list
Path                Type          Accessor                    Description                Version
----                ----          --------                    -----------                -------
production-copy/    kubernetes    auth_kubernetes_f09ee9e8    n/a                        n/a
token/              token         auth_token_517e4e2e         token based credentials    n/a

$ vault read auth/production-copy/config
Key                       Value
---                       -----
disable_iss_validation    true
disable_local_ca_jwt      true
issuer                    n/a
kubernetes_ca_cert        -----BEGIN CERTIFICATE-----
something
-----END CERTIFICATE-----
kubernetes_host           https://some-kubernetes-cluster.com
pem_keys                  []

$ vault list auth/production-copy/role
Keys
----
default

$ vault read auth/production-copy/role/default
Key                                 Value
---                                 -----
alias_name_source                   serviceaccount_uid
bound_service_account_names         [default]
bound_service_account_namespaces    [default]
policies                            [allow_secrets]
token_bound_cidrs                   []
token_explicit_max_ttl              0s
token_max_ttl                       0s
token_no_default_policy             false
token_num_uses                      0
token_period                        0s
token_policies                      [allow_secrets]
token_ttl                           1h
token_type                          default
ttl                                 1h
```

Everything looks good! :D

Note that, the destination Vault does not have any user defined Vault policies. And as mentioned before in the Notes, the policies referred to in the Kubernetes Auth Method Roles (Roles Configuration) in the source Vault are NOT copied to the destination Vault.

In this case, the `allow_secrets` Vault policy referred to in the `default` Kubernetes Auth Method Role is not present in the destination Vault

```bash
$ vault policy list
default
root
```

This alone, one needs to create themselves manually. In the future, one can use a separate tool for copying Vault Policies. I also plan to add the feature to this tool to copy just the Vault Policies referred to in the Kubernetes Auth Method Roles (Roles Configuration) from source Vault to destination Vault :) So that this manual step is NOT required :)

Note: If the destination Vault does NOT have the Kubernetes Auth Method enabled at the given destination mount path, then the tool throws an error similar to this -

```bash
$ ./vault-k8s-auth-cp production production

Source Kubernetes Auth Config: map[disable_iss_validation:true disable_local_ca_jwt:true issuer: kubernetes_ca_cert:-----BEGIN CERTIFICATE-----
something
-----END CERTIFICATE-----
 kubernetes_host:https://some-kubernetes-cluster.com pem_keys:[]]
Error writing k8s auth config to destination vault: Error making API request.

URL: PUT http://127.0.0.1:8300/v1/auth/production/config
Code: 404. Errors:

* no handler for route "auth/production/config". route entry not found.
```

Note: If the Vault Token / Credentials used for the destination Vault is not valid / wrong / does not have enough access, then the tool throws an error similar to this regardless of if the Kubernetes Auth Method is enabled or not at the given destination mount path -

For example, for a destination mount path that is present with Kubernetes Auth Method enabled in it -

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="some-big-token-here"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="blah" # wrong Vault Token

$ ./vault-k8s-auth-cp production production

Source Kubernetes Auth Config: map[disable_iss_validation:true disable_local_ca_jwt:true issuer: kubernetes_ca_cert:-----BEGIN CERTIFICATE-----
something
-----END CERTIFICATE-----
 kubernetes_host:https://some-kubernetes-cluster.com pem_keys:[]]
Error writing k8s auth config to destination vault: Error making API request.

URL: PUT http://127.0.0.1:8300/v1/auth/production/config
Code: 403. Errors:

* permission denied
```

For a destination mount path that is NOT present at all in the destination Vault, for that also same error with bad Vault Token / Credentials -

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="some-big-token-here"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="blah" # wrong Vault Token

$ ./vault-k8s-auth-cp production production-copy

Source Kubernetes Auth Config: map[disable_iss_validation:true disable_local_ca_jwt:true issuer: kubernetes_ca_cert:-----BEGIN CERTIFICATE-----
something
-----END CERTIFICATE-----
 kubernetes_host:https://some-kubernetes-cluster.com pem_keys:[]]
Error writing k8s auth config to destination vault: Error making API request.

URL: PUT http://127.0.0.1:8300/v1/auth/production-copy/config
Code: 403. Errors:

* permission denied
```
