# kpv - KeePass Vault CLI

kpv is a CLI for automating secrets management with KeePass (`.kdbx`) vaults. It supports initializing vaults, getting/setting secrets, listing, importing, exporting, removing, ensuring, and syncing entries, with sensible defaults and OS keyring integration.

## Install

Build locally:

```bash
mise r build
./bin/kpv --help
```

## Quick Start

Initialize the default vault and auto-generate a password:

```bash
kpv init
```

This creates `~/.local/share/kpv/default.kdbx` on Linux/macOS (or `%LOCALAPPDATA%/kpv/default.kdbx` on Windows). The generated password is saved to your OS keyring and a `.key` file in `~/.local/share/kpv/` for backup.

Set a secret:

```bash
kpv secrets set --key api-token --value "super-secret"
```

Get a secret:

```bash
kpv secrets get --key api-token
```

List secrets:

```bash
kpv secrets ls
kpv secrets ls "app-*"
```

## Global Flags

Available on all commands (env vars in parentheses):
- `-V, --vault` (`KPV_VAULT`): Vault name or path.
- `-P, --vault-password` (`KPV_PASSWORD`): KeePass vault password.
- `-F, --vault-password-file` (`KPV_PASSWORD_FILE`): Path to file containing the KeePass vault password.
- `--vault-os-secret` (`KPV_VAULT_OS_SECRET`): OS secret store key used to retrieve the vault password.

Passwords can also be read from the OS keyring (if previously saved) or a `.key` file in the kpv data directory.

## Commands

### `kpv init`

Initialize a new KeePass vault.

Examples:
- Default vault: `kpv init`
- Named vault in global location: `kpv init --vault myvault --global`
- Specific path: `kpv init --vault /path/to/vault.kdbx`
- Custom password: `kpv init --vault myvault --vault-password "my-secure-password"`

Notes:
- If no password is provided, a strong one is generated and saved to the OS keyring and a `.key` file.
- Prevents overwriting an existing file.

Flags:
- `-V, --vault` Vault name or path
- `--global` Place simple `--vault` names under the default directory
- `-P, --vault-password` Vault password
- `-F, --vault-password-file` Path to file containing the vault password
- `--vault-os-secret` OS secret store key used to retrieve the vault password

### `kpv secrets ensure`

Get a secret if it exists; otherwise generate and set it, then output the value.

Examples:
- `kpv secrets ensure --key my-secret`
- `kpv secrets ensure --key my-secret --size 32 -U -S`
- `kpv secrets ensure my-secret`

Generation options:
- `--size` (default 16) - sets the length of the generated secret. 
- `-U, --no-upper` - disables using upper english characters.
- `-L, --no-lower` - disables using lower english characters.
- `-D, --no-digits` - disables using digits.
- `-S, --no-special` - disables special characters/symbols.
- `--special <special>` - set the special symbols that are used.
- `--chars <chars>` - mutually exclusive with the other character options.

### `kpv secrets get`

Get one or more secrets by key.

Examples:
- Single: `kpv secrets get --key my-secret`
- Multiple: `kpv secrets get --key secret1 --key secret2`
- Formats: `json`, `sh|bash|zsh`, `pwsh|powershell`, `dotenv`, `null`, 
  `azure-devops`, `github-actions`, `run`, and the default `text`.

Flags:
- `-k, --key` Secret key (repeatable)
- `-f, --format` Output format

### `kpv secrets get-string`

Get a custom string field from an entry.

Example:
- `kpv secrets get-string --key my-secret --field api-key`

Notes:
- Standard fields (`Title`, `Username`, `Password`, `URL`, `Notes`) should be retrieved via `secrets get`.

Flags:
- `-k, --key` Entry title
- `-f, --field` Custom field name

### `kpv secrets set`

Create or update a secret value.

Input methods (mutually exclusive; if none given, `--generate` is assumed):
- `--value` - sets the value on the command line. Should only be used by a script. 
- `--file` - sets the value using the file provided
- `--env` - sets the value using the environment variable given.
- `--stdin` - sets the value using Standard Input.
- `--generate` - generate the value using a cryptographically secure secret generator.

Generation options:
- `--size` (default 16), 
- `-U, --no-upper` - disables using upper english characters.
- `-L, --no-lower` - disables using lower english characters.
- `-D, --no-digits` - disables using digits.
- `-S, --no-special` - disables special characters/symbols.
- `--special <special>` - set the special symbols that are used.
- `--chars <chars>` - mutually exclusive with the other character options.


Examples:
- `kpv secrets set --key my-secret --value "secret-value"`
- `kpv secrets set --key my-secret --file ./secret.txt`
- `echo "secret" | kpv secrets set --key my-secret --stdin`
- `kpv secrets set --key my-secret --generate --size 32`

### `kpv secrets set-string`

Set a custom string field on an entry (protected or unprotected).

Examples:
- `kpv secrets set-string --key my-secret --field api-key --value "abc123"`
- `kpv secrets set-string --key my-secret --field cert --file cert.pem`
- `echo "value" | kpv secrets set-string --key my-secret --field custom --stdin`
- Protected: `kpv secrets set-string --key my-secret --field token --value "xyz" --protected`

Flags:
- `-k, --key` - The full path to the entry including groups using forward
  slashes e.g. group1/next-group/entry-name
- `-f, --field` - The custom string field name.
- `-v, --value` - The value. This should only be used in a script.
- `--file` - Sets the value by reading it from the given file.
- `--env`  - Sets the value using the given environment variable
- `--stdin` - Sets the value using the Standard Input. Useful for piping the value.
- `-p, --protected` - Instructs keepass to protect the value by encrypting it.

### `kpv secrets ls`

List secrets. Supports glob filters.

Examples:
- `kpv secrets ls`
- `kpv secrets ls "app-*"`
- `kpv secrets list "db-*-password"`

### `kpv secrets rm`

Remove one or more secrets.

Examples:
- `kpv secrets rm --key secret1 --key secret2`
- `kpv secrets rm --key my-secret --yes`

Flags:
- `-k, --key` Secret key(s)
- `-y, --yes` Skip confirmation

### `kpv secrets export`

Export all secrets to JSON.

Examples:
- `kpv secrets export --json --file secrets.json`
- `kpv secrets export --json --pretty`

Flags:
- `--json` Required for now
- `-f, --file` Output file (default stdout)
- `--pretty` Pretty-print JSON

Output per secret:
- `value`, optional `username`, `url`, `notes`
- `strings`: `{ value, encrypted }`
- `tags`: map of tag names (empty values)

### `kpv secrets import`

Import secrets from JSON.

Examples:
- `kpv secrets import --json --file secrets.json`
- `cat secrets.json | kpv secrets import --json --stdin`

Flags:
- `--json` Required for now
- `-f, --file` Input file
- `--stdin` Read from stdin

Accepted JSON per secret:
- Simple string value, or an object with: `value`, `username`, `url`, `notes`, `strings` (`{ value, encrypted }`), optional generation: `ensure`, `size`, `noUpper`, `noLower`, `noDigits`, `noSpecial`, `special`, `chars`.

### `kpv secrets sync`

Sync secrets from JSON, updating only when values differ; supports delete, ensure, dry-run.

Examples:
- `kpv secrets sync --json --file secrets.json`
- `cat secrets.json | kpv secrets sync --json --stdin`
- `kpv secrets sync --json --file secrets.json --dry-run`

Flags:
- `--json` Required for now
- `-f, --file` Input file
- `--stdin` Read from stdin
- `--dry-run` Show changes without applying

JSON per secret supports: `value`, `username`, `url`, `notes`, `strings { value, encrypted }`, `tags { name }`, `delete`, `ensure`, `size`, `noUpper`, `noLower`, `noDigits`, `noSpecial`, `special`, `chars`.

### `kpv config`

Manage CLI configuration values.

Subcommands:
- `kpv config get <key>`: Print a config value.
- `kpv config set <key> <value>`: Set a config value.
- `kpv config ls [filter]`: List keys (optional glob filter).
- `kpv config rm <key>`: Remove a config value.

Examples:
- `kpv config set defaults.path /path/to/default.kdbx`
- `kpv config get defaults.path`
- `kpv config ls "alias*"`

### `kpv config aliases`

Manage KeePass path aliases stored in config.

Subcommands:
- `kpv config aliases set <name> <path>`
- `kpv config aliases get <name>`
- `kpv config aliases ls [filter]`

### `kpv config secret`

Manage KeePass database passwords in the OS secret store.

Subcommands:
- `kpv config secret get <path|file:///path|kpv:///path>`: Retrieve from OS keyring.
- `kpv config secret set <path> [secret]`: Store to OS keyring; supports `--file`, `--stdin`, `--env`, `--prompt`.

Notes:
- Paths are normalized to `kpv:///absolute/path`.
- Combine with global `--vault-os-secret` for retrieval during CLI runs.

## Defaults & Discovery

- Default vault path (Linux/macOS): `~/.local/share/kpv/default.kdbx`.
- Simple names resolve under the default directory when used with `init --global`.
- Password resolution order: `--password` > `--password-file` > env (`KPV_PASSWORD`/`KPV_PASSWORD_FILE`) > OS keyring > `.key` file.

## Environment Variables

- `KPV_VAULT` default vault name/path
- `KPV_PASSWORD` vault password
- `KPV_PASSWORD_FILE` path to a file with the password
- `KPV_VAULT_OS_SECRET` OS secret store key for retrieving the vault password

## Tips

- Back up generated `.key` files securely; consider removing them after storing elsewhere.
- Prefer OS keyring storage for convenience and security.

## Help

Use `kpv --help` or any subcommand with `--help` for details.