# Pin Deployment Images in your ClusterServiceVersion

The `pin-deploy-imgs-in-csv` tool simply reads your ClusterServiceVersion YAML
manifest on disk, and replaces any image references in deployment containers
that use a tag with the corresponding digest.

This tool may also add some additional fields with null or empty values if the
[ClusterServiceVersion](https://github.com/operator-framework/api/blob/master/pkg/operators/v1alpha1/clusterserviceversion_types.go#L578)
type does not omit them.

These additions should be harmless (ex. `CreationTimestamp: null`).

This will overwrite your ClusterServiceVersion file provided as input.

## Usage

```text
$ pin-deploy-imgs-in-csv help
pin-deploy-imgs-in-csv /path/to/clusterserviceversion.yaml

This tool will check your ClusterServiceVersion's
deployment containers for images referenced using a tag, and replace
the tag with the digest of the image at that point in time
```

To see the release version, run `pin-deploy-imgs-in-csv version`.

## Testing

A basic test suite is available using `make test`.

```text
$ make test
./test/test.sh
Checking system for requirements.
Verifying the test binary is built and lives at a known path.
Creating a temporary directory and copying test fixtures into it.
Running the pin tool against the fixture.
Test: the tool must not modify any images that are already referenced via digest.
Test: the tool must resolve an images digest when it does not contain a tag at all.
Test: the tool must resolve an image that is referenced via tag to its digest, and inject the tag as a comment.
All tests passed!
```
