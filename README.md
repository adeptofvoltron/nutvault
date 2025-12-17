# nutvault

![nutvault logo](notvault_logo.png)

## Usage

nutvault is a CLI tool for managing encrypted environment variable vaults. It provides secure storage and retrieval of environment variables from `.env` files.

### Commands

#### Collect

Collect variables from a `.env` file and save them to a vault project.

```bash
# Collect from default .env file with default key
nutvault collect myproject

# Collect from custom .env file
nutvault collect myproject --env-file .env.production

# Collect with custom key file
nutvault collect myproject --key-file ~/.nutvault/mykey.hex
```

The `collect` command reads all variables from a `.env` file and saves them to an encrypted vault project. The project is stored at `~/.nutvault/projects/<projectName>.<hash>`.

#### Fill

Fill empty variables in a `.env` file with values from the vault.

```bash
# Fill empty variables in default .env file
nutvault fill myproject

# Fill empty variables in custom .env file
nutvault fill myproject --env-file .env.production

# Fill with custom key file
nutvault fill myproject --key-file ~/.nutvault/mykey.hex
```

The `fill` command reads variables from a vault project and fills only empty variables in a `.env` file. Variables that already have values are not modified.

#### Swap

Replace all variable values in a `.env` file with values from the vault.

```bash
# Swap all variables in default .env file
nutvault swap myproject

# Swap all variables in custom .env file
nutvault swap myproject --env-file .env.production

# Swap with custom key file
nutvault swap myproject --key-file ~/.nutvault/mykey.hex
```

The `swap` command reads variables from a vault project and replaces all variable values in a `.env` file. All existing variable values will be overwritten with values from the vault.

#### Remove

Delete a vault project and all its contents.

```bash
# Remove project with default key
nutvault remove myproject

# Remove project with custom key file
nutvault remove myproject --key-file ~/.nutvault/mykey.hex
```

The `remove` command deletes an entire vault project and all its contents. This operation cannot be undone.

#### List

List all vault projects.

```bash
# List all projects
nutvault list
```

The `list` command displays all vault projects stored in `~/.nutvault/projects/`. For each project, it shows the project name, hash, path, and number of variables.

### Options

- `--env-file, -e`: Path to `.env` file (default: `.env` in current directory)
- `--key-file, -k`: Path to key file in hex format (default: use default user key)

### Key Files

If no key file is specified, nutvault uses a default key generated deterministically from your user and host information. For custom encryption keys, provide a key file containing exactly 64 hex characters (32 bytes).

