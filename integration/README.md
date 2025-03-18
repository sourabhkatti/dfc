## Integration testing

These tests verify that converted Dockerfiles actually build.

In [`../testdata/`](../testdata/) , if we see a folder called `<testname>-integration>`, this automatically opts the test into
integration testing.

To run:

```sh
go test -v ./integration/ -tags=integration
```

TODO: hook these tests up to CI, requires a valid Chainguard org
with access to all necessary base images.
