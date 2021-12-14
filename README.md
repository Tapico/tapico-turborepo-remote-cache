# Company Turborepo Remote Cache

This is an implementation of Vercel's Turborepo Remote Cache API endpointss used
by the `turborepo` CLI command. This solution allows you to get control over were
the cache artifacts are being stored.

The CLI tool currently supports the following targets for the cache artificats:

  - `gcs`: Google Cloud Storage
  - `s3`: Amazon S3
  - `local`: The local file system

## Acknowledgments

Thank you to the developers of the libraries used by this application, especially
the authors of the `stow` and the `opentelemetry-go` libraries
