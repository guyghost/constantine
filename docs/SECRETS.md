# Secrets Management with 1Password

The bot expects all credentials to arrive via environment variables. To avoid
storing raw keys on disk, leverage the 1Password CLI (`op`) to inject secrets
at runtime.

## 1. Create vault items

1. Create a dedicated vault (e.g. `Trading`).
2. Add items for each exchange:
   - **Hyperliquid** with fields `API Key`, `API Secret`.
   - **Coinbase** with fields `API Key`, `API Secret`, optional `Portfolio ID`.
   - **dYdX** with either `Mnemonic` or API credentials plus `Wallet Address`.
3. Restrict access to the vault (per-user or service account).

## 2. Map fields in `.env.op.template`

The repository ships with `.env.op.template` containing 1Password secret
references (e.g. `op://Trading/Hyperliquid/API Secret`). Update the vault name
or field names if they differ from your setup.

## 3. Run commands with secrets injected

Use `op run --env-file .env.op.template -- <command>` to populate the
environment on the fly without creating a physical `.env` file:

```bash
# Authenticate with op beforehand: `eval "$(op account add ...)"` and `op signin`
op run --env-file .env.op.template -- make run
```

Any command (`make test`, `go run`, CI jobs, etc.) can be wrapped the same way.
The variables are available to the process but never written to disk.

## 4. Optional: materialize a local `.env`

If you absolutely need a local `.env` (for example to debug tools that do not
support `op run`), create it transiently with `op inject` and remove it when
you are done:

```bash
op inject -i .env.op.template -o .env.local
# use .env.local carefully, then delete it
rm .env.local
```

## 5. CI/CD integration

Service accounts can fetch the same secrets by running `op run --env-file ...`
in the pipeline. Store the 1Password token in the CI secret store and avoid
printing the resulting environment.

Following this workflow keeps the `internal/config` validation satisfied while
ensuring sensitive API keys live only inside 1Password.
