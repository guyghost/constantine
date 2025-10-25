# CI Integration

This project relies on environment variables provided by `.env.op.template`.
To keep secrets inside 1Password even in CI, wrap every sensitive command with
`op run --env-file .env.op.template -- <command>`.

## GitHub Actions example

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    env:
      OP_SERVICE_ACCOUNT_TOKEN: ${{ secrets.OP_SERVICE_ACCOUNT_TOKEN }}
    steps:
      - uses: actions/checkout@v4
      - uses: 1password/load-secrets-action@v1
        with:
          export-env: false
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Run tests with secrets
        run: |
          op run --env-file .env.op.template -- go test ./...
```

Replace `OP_SERVICE_ACCOUNT_TOKEN` with the token generated for the service
account that has access to the vault items referenced in `.env.op.template`.

## Local automation scripts

Any automation script (Makefile target, shell script) can be prefixed the same
way:

```bash
op run --env-file .env.op.template -- make run
```

This keeps developer and CI workflows aligned while ensuring the sensitive
values never touch disk.
