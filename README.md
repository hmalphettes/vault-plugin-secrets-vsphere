# Vault vSphere Secrets Plugin

VSphere Secrets plugins is a secrets engine plugin for [HashiCorp Vault](https://www.vaultproject.io/). It is meant for demonstration purposes only and should not be used in production at this time.

It was forked from the Mock plugin found in the vault-guides.

Most of the code is modeled after the github.com/hashicorp/vault-plugin-secrets-azure

The license is identical to that of the vault-plugin-secrets-azure: MPL-2.0.

## Roadmap

Exploration and POC stage.

Store a vSphere `sudoer` user in Vault.
Configure roles that create short lived users or sessions managed by Vault.
Map existing users to roles.
Put in place unit tests relying on govmomi/vcsim's mock framework for vSphere.

## Setup

Most secrets engines must be configured in advance before they can perform their
functions. These steps are usually completed by an operator or configuration
management tool.

1. Enable the vSphere secrets engine:

    ```bash
    $ vault secrets enable vsphere
    Success! Enabled the vsphere secrets engine at: vsphere/
    ```

    By default, the secrets engine will mount at the name of the engine. To
    enable the secrets engine at a different path, use the `-path` argument.

1. Configure the secrets engine with `admin` account credentials:

    ```bash
    $ vault write vsphere/config \
    url=$GOVMOMI_URL \
    username=$GOVMOMI_USERNAME \
    password=$GOVMOMI_PASSWORD \
    insecure=$GOVMOMI_INSECURE

    Success! Data written to: vsphere/config
    ```

    Note that it is not required to provide an admin account at all.

    In that case only roles configured with an existing user and password will be functional

2. Configure a role. A role may be set up with either an existing user, or
a set of vSphere roles that will be assigned to a dynamically created service principal.

To configure a role called "my-role" with an existing user:

    ```bash
    $ vault write vsphere/roles/my-role username=<existing_username> password=<existing_password-or-empty> ttl=1h
    ```

Alternatively, to configure the role to create a new user with vSphere roles (?? does this exist ??):

    ```bash
    $ vault write vsphere/roles/my-role ttl=1h vsphere_roles=VMsAdmin,DisksAdmin" vsphere_groups="PerfView"
    ```

Roles may also have their own TTL configuration that is separate from the mount's
TTL. For more information on roles see the [roles](#roles) section below.



## Usage

All commands can be run using the provided [Makefile](./Makefile). However, it may be instructive to look at the commands to gain a greater understanding of how Vault registers plugins. Using the Makefile will result in running the Vault server in `dev` mode. Do not run Vault in `dev` mode in production. The `dev` server allows you to configure the plugin directory as a flag, and automatically registers plugin binaries in that directory. In production, plugin binaries must be manually registered.

This will build the plugin binary and start the Vault dev server:
```bash
# Build vSphere plugin and start Vault dev server with plugin automatically registered
$ make
```

Now open a new terminal window and run the following commands:
```bash
# Open a new terminal window and export Vault dev server http address
$ export VAULT_ADDR="http://127.0.0.1:8200"

# Enable the vSphere secrets plugin
$ vault secrets enable vsphere

# Configure the vSphere secrets engine
$ vault write vsphere/config url="http://localhost:8056" username="root" password="root" insecure="true"
Success! Data written to: vsphere/config

# Retrieve secret from Mock secrets engine
$ vault read mock/test
Key      Value
---      -----
hello    world
```

## License

Mock was contributed to the HashiCorp community by [hasheddan](https://github.com/hasheddan/vault-plugin-secrets-covert). In doing so, the original license has been removed.

Most of the code is modeled after the github.com/hashicorp/vault-plugin-secrets-azure

The license is identical to that of the vault-plugin-secrets-azure: MPL-2.0.
