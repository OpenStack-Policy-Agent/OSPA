## Testing

### pkg/ unit tests (CI default)

```bash
go test ./pkg/... -count=1
```

### pkg/ unit coverage

```bash
go test ./pkg/... -count=1 -race -coverprofile=pkg.cover.out
go tool cover -func=pkg.cover.out
go tool cover -html=pkg.cover.out -o pkg.cover.html
```

### pkg/ integration tests (optional)

Integration tests are behind a build tag and will **skip** unless OpenStack config is present.

```bash
go test -tags=integration ./pkg/... -count=1
```

### Convenience script

```bash
./scripts/test-pkg.sh
RUN_INTEGRATION=1 ./scripts/test-pkg.sh
```


