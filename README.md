# Tapico Turborepo Remote Cache

This is an implementation of Vercel's Turborepo Remote Cache API endpointss used
by the `turborepo` CLI command. This solution allows you to get control over were
the cache artifacts are being stored.

The CLI tool currently supports the following targets for the cache artificats:

  - `gcs`: Google Cloud Storage
  - `s3`: Amazon S3
  - `local`: The local file system

## Running the application

You can execute this application by running:

```bash
./tapico-turborepo-remote-cache
```

This will start a web-server on port `8080`

### Configuration

The server supports three kind of cloud storage, which are `s3`, `gcs` and `local`,
the latter will store the cache artefects on the local file system on a relative path.

The configuration is currently handled via environment variables, the following
are available:

  - CLOUD_PROVIDER_KIND: `s3`, `gcs` or `local`
  - GOOGLE_CREDENTIALS_FILE: location the google credentials json file
  - GOOGLE_PROJECT_ID: the project id

  - AWS_ENDPOINT: the endpoint to connect to for Amazon S3
  - AWS_ACCESS_KEY_ID: the Amazon acces key id
  - AWS_SECRET_ACCESS_KEY: the Amazon secret access key
  - AWS_S3_REGION_NAME: the region for Amazon S3
  - CLOUD_SECURE: whether the endpoint is secure (https) or not, can be `true` or `false`
  - CLOUD_FILESYSTEM_PATH: the relative path to the file system

## Developing

In the `dev` directory you can find a docker compose file which starts, a Minio
S3 compatible service, for testing the Amazon S3 integration, the path for this
is: http://127.0.0.1:9000

Another service running is a fake Google Cloud Storage server on port
http://127.0.0.1:9100. If you want to use this you need to make sure you set the
following environment variable:

```bash
export STORAGE_EMULATOR_HOST=http://localhost:9100
```

The `STORAGE_EMULATOR_HOST` is used to activate a special code path in
the Google Cloud Storage library for Go.

## Acknowledgments

Thank you to the developers of the libraries used by this application, especially
the authors of the `stow` and the `opentelemetry-go` libraries
