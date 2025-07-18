# Developing `libmodal`

## modal-go development

Clone this repository. You should be all set to run an example.

```bash
go run ./examples/sandbox
```

Whenever you need a new version of the protobufs, you can regenerate them:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
scripts/gen-proto.sh
```

We check the generated into Git so that the package can be installed with `go get`.

## modal-js development

Setup after cloning the repo with submodules:

```bash
npm install
```

Then run a script with:

```bash
node --import tsx path/to/script.ts
```

### gRPC support

We're using `nice-grpc` because the `@grpc/grpc-js` library doesn't support promises and is difficult to customize with types.

This gRPC library depends on the `protobuf-ts` package, which is not compatible with tree shaking because `ModalClientDefinition` transitively references every type. However, since `modal-js` is a server-side package, having a larger bundled library is not a huge issue.

## How to publish

1. Update `CHANGELOG.md`. Make sure that you're on a clean commit, then run the following update the `modal-js` version and update the changelog:

```bash
python ci/release.py version patch  # or 'minor'
```

2. Push changes to `main` or open a PR.

3. Publish both `modal-js` and `modal-go`:

```bash
python ci/release.py publish
```
