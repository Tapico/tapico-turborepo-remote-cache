# Tapico Turborepo Remote Cache

This is an implementation of Vercel's Turborepo Remote Cache API endpoints used
by the `turborepo` CLI command. This solution allows you to get control over where
the cache arteficats are being stored.

The CLI tool currently supports the following targets for the cache arteficats:

  - `gcs`: Google Cloud Storage
  - `s3`: Amazon S3
  - `local`: The local file system

## Running the application

You can execute this application by running when you want to store your cache artefacts
on a Amazon S3 compatible cloud storage provider, it will start a HTTP server on port 8080:

```bash
./tapico-turborepo-remote-cache \
  --kind="s3" \
  --s3.endpoint="http://127.0.0.1:9000" \
  --s3.accessKeyId="minio" \
  --s3.secretKey="miniosecretkey" \
  --s3.region="eu-west-1" \
  --turbo-token="your-turbo-token"
```

*Note*: The above example can be used to test against the Minio instance of the `docker-compose.yml` file found in the `dev`-directory.

At this time the server doesn't support running over HTTPS, you might want to consider
using a load balancer to expose the server over HTTPS to the internet.

You can download [binaries](https://github.com/Tapico/tapico-turborepo-remote-cache/releases) of the applications via the Releases page, and pre-build docker images are available in the [`tapico-turborepo-remote-cache` section in Packages](https://github.com/orgs/Tapico/packages/container/package/tapico-turborepo-remote-cache) section.

## Example

An example is available that demonstrate how to use the the Turbo remote cache server together with the [Amazon S3 compatible object atorage Minio](https://min.io) using
Docker Compose, the example can be found in [examples/with-docker-compose](https://github.com/Tapico/tapico-turborepo-remote-cache/tree/main/examples/with-docker-compose).

### Configuration

The server supports three kind of cloud storage, which are `s3`, `gcs` and `local`,
the latter will store the cache artefects on the local file system on a relative path.

The configuration is currently handled via environment variables, the following
are available:

  - `CLOUD_PROVIDER_KIND`: `s3`, `gcs` or `local`
  - `BUCKET_NAME`: the name of the bucket to store the cache artefacts
  - `LISTEN_ADDRESS`: the address the server to listen to (defaults to: `127.0.0.1:8080`) 
     when deploying it to the internet you should consider using `0.0.0.0:8080` instead, the `8080` representa the port.
  - `GOOGLE_CREDENTIALS_FILE`: location the google credentials json file
  - `GOOGLE_PROJECT_ID`: the project id
  - `GOOGLE_ENDPOINT`: the endpoint to use for Google Cloud Storage (e.g. for emulator)
  - `AWS_ENDPOINT`: the endpoint to connect to for Amazon S3
  - `AWS_ACCESS_KEY_ID`: the Amazon acces key id
  - `AWS_SECRET_ACCESS_KEY`: the Amazon secret access key
  - `AWS_S3_REGION_NAME`: the region for Amazon S3
  - `CLOUD_SECURE`: whether the endpoint is secure (https) or not, can be `true` or `false`
  - `CLOUD_FILESYSTEM_PATH`: the relative path to the file system
  - `TURBO_TOKEN`: comma seperated list of accepted TURBO_TOKENS

Alternatively, you can also use the CLI arguments:

```bash
usage: tapico-turborepo-remote-cache --turbo-token=TURBO-TOKEN [<flags>]

A tool to work with Vercel Turborepo to upload/retrieve cache artefacts to/from popular cloud providers

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose                  Verbose mode.
      --kind="s3"                Kind of storage provider to use (s3, gcs, local). ($CLOUD_PROVIDER_KIND)
      --secure                   Enable secure access (or HTTPs endpoints).
      --bucket="tapico-remote-cache"
                                 The name of the bucket ($BUCKET_NAME)
      --enable-bucket-per-team   The name of the bucket
      --turbo-token=TURBO-TOKEN  The comma separated list of TURBO_TOKEN that the server should accept ($TURBO_TOKEN)
      --google.endpoint=GOOGLE.ENDPOINT
                                 API Endpoint of cloud storage provide to use ($GOOGLE_ENDPOINT)
      --google.project-id=GOOGLE.PROJECT-ID
                                 The project id relevant for Google Cloud Storage ($GOOGLE_PROJECT_ID).
      --google.credentials=GOOGLE.CREDENTIALS
                                 The path to the credentials file ($GOOGLE_APPLICATION_CREDENTIALS).
      --local.project-id=LOCAL.PROJECT-ID
                                 The relative path to storage the cache artefacts when 'local' is enabled ($CLOUD_FILESYSTEM_PATH).
      --s3.endpoint=S3.ENDPOINT  The endpoint to use to connect to a Amazon S3 compatible cloud storage provider ($AWS_ENDPOINT).
      --s3.accessKeyId=S3.ACCESSKEYID
                                 The Amazon S3 Access Key Id ($AWS_ACCESS_KEY_ID).
      --s3.secretKey=S3.SECRETKEY
                                 The Amazon S3 secret key ($AWS_SECRET_ACCESS_KEY).
      --s3.region=S3.REGION      The Amazon S3 region($AWS_S3_REGION_NAME).
```

*Note*: You can use the environment variable `LISTEN_ADDRESS` to control to the address
the web-server should listen to. The default value currently is `127.0.0.1:8080`.

*Note*: Most of the arguments also alternatively accepts a environment variable instead
the environment name is between brackets. For example, if you want to specify the list
of accepted Turborepo tokens you can also set the environment variable `TURBO_TOKEN` instead
of using the `--turbo-token`-argument.

## Storing cache artefacts

The service allows to store cache artefacts into Amazon S3 compatible cloud storage, or
Google Cloud Storage. If the option `--enable-bucket-per-team` is enabled, the service
will try to create a new bucket for each team id that's received.

Alternatively, you can also use a single bucket, the name of the bucket can be controlled through
the `--bucket` option. Using this approach does mean that each of the passed team id's will
become a subdirectory in the bucket, and the directory will contain all the cache artefacts
uploaded by Turborepo.

## Running the server

Two approaches are available to run the Tapico Turborepo Remote cache solution,
one is to use the docker image that is available on [Github Packages](https://github.com/Tapico/tapico-turborepo-remote-cache/pkgs/container/tapico-turborepo-remote-cache) at:
https://github.com/Tapico/tapico-turborepo-remote-cache/pkgs/container/tapico-turborepo-remote-cache
the second solution is to download the binary via the [Github Releases](https://github.com/Tapico/tapico-turborepo-remote-cache/releases) at:
https://github.com/Tapico/tapico-turborepo-remote-cache/releases

You can deploy the service in your preferred way, we have been successfully running it via
Google Cloud Run, but via Kubernetes or any solution that accepts a docker image or
Go binary should work.

### Google Cloud Run

Before you can run the service on Cloud Run, you need to make sure that you have
the docker image available in the Google Container Registry or Artefacts registry.
If you are authenticated to the registry of choice, you can tag the docker image
and push it to the registry.

For the artefacts registry the following works:
```bash
# Pull the docker image from Github Packages you want to use via:
docker pull ghcr.io/tapico/tapico-turborepo-remote-cache:sha-9229998

# Tag the docker image for the container registry, e.g.
docker tag ghcr.io/tapico/tapico-turborepo-remote-cache:sha-9229998 europe-west2-docker.pkg.dev/dev-sandbox-20211210/devops/tapico-remote-cache
# Note: You can find the `pkg.dev url you need to use in Google Cloud Console

# Last step, is to push the docker image to the container registry
docker push europe-west2-docker.pkg.dev/dev-sandbox-20211210/devops/tapico-remote-cache
```

After the Docker image is available, you can create a new deployment in Cloud Run and
select the docker image that you pushed to the Google container registry. You can configure
the deployment to define the appropriate environment variables or specify the
run arguments of your choice.

Currently, you need to also set a special environment variable that defines the
address the server should listen to for incoming requests to be `0.0.0.0:8080` to
ensure the Tapico Turborepo Remote cache can receive requests on Cloud Run.

## Configuring Turbo

After you have started the server you need to change the configuration of Turbo
to ensure it's pointing to our server for the API server. Currently, any of the
login functionality is not implemented. You can adapt the `.turbo/config.json`-file
in the root of your mono repo.

NOTE: You **don't** have to run `turbo login` step.

```json
{
  "teamId": "team_blah",
  "apiUrl": "http://127.0.0.1:8080"
}
```

NOTE: `teamId` must start with "team_".

After this you should be able to run `turbo` e.g. `turbo run build --force` to
force the generating of new cache artefacts and upload it to our server.

Alternatively, you can also use the arguments `--api="http://127.0.0.1:8080" --token="xxxxxxxxxxxxxxxxx"`

Token can also be specified as environment variable `TURBO_TOKEN=xxxxx`, for example: `TURBO_TOKEN=xxxx turbo run build`

If the option `--enable-bucket-per-team` is enabled, the `teamId` in `.turbo/config.json` is
used to generate a bucket in the cloud storage provider, as the id might be an invalid name
the team identifier a MD5 hash is generated and used as the bucket name.

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

*Tip*: If the Remote Cache is not working as expected, you can use an application
like ProxyMan and force `turbo` CLI the application's HTTP proxy so you can get
insight in the outgoing HTTP requests. To do this, you can run `turbo` the following
way `HTTP_PROXY=192.168.1.98:9090 turbo run build`.

You might need to use `HTTPS_PROXY` instead when the API server location is running
over HTTPS instead of HTTP.

## Acknowledgments

Thank you to the developers of the libraries used by this application, especially
the authors of the `stow`, `mux`, `kingpin`, and the `opentelemetry-go` libraries.

Thank you to Jared Palmer for building the awesome Turborepo CLI tool!
