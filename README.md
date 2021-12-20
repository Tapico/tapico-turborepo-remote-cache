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
./tapico-turborepo-remote-cache --kind="s3" --s3.endpoint="http://127.0.0.1:9000" --s3.accessKeyId="minio" --s3.secretKey="miniosecretkey" --s3.region="eu-west-1" --turbo-token="your-turbo-token"
```
*Note*: The above example can be used to test against the Minio instance of the `docker-compose.yml` file found in the `dev`-directory.

At this time the server doesn't support running over HTTPS, you might want to consider
using a load balancer to expose the server over HTTPS to the internet.

You can download [binaries](https://github.com/Tapico/tapico-turborepo-remote-cache/releases) of the applications via the Releases page, and pre-build docker images are available in the [`tapico-turborepo-remote-cache` section in Packages](https://github.com/orgs/Tapico/packages/container/package/tapico-turborepo-remote-cache) section.

### Configuration

The server supports three kind of cloud storage, which are `s3`, `gcs` and `local`,
the latter will store the cache artefects on the local file system on a relative path.

The configuration is currently handled via environment variables, the following
are available:

  - `CLOUD_PROVIDER_KIND`: `s3`, `gcs` or `local`
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

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose                  Verbose mode.
      --kind="s3"                Kind of storage provider to use (s3, gcp, local). ($CLOUD_PROVIDER_KIND)
      --secure                   Enable secure access (or HTTPs endpoints).
      --turbo-token=TURBO-TOKEN  The comma separated list of TURBO_TOKEN that the server should accept ($TURBO_TOKEN)
      --google.endpoint="http://127.0.0.1:9000"
                                 API Endpoint of cloud storage provide to use ($GOOGLE_ENDPOINT)
      --google.project-id=GOOGLE.PROJECT-ID
                                 The project id relevant for Google Cloud Storage ($GOOGLE_PROJECT_ID).
      --local.project-id=LOCAL.PROJECT-ID
                                 The relative path to storage the cache artefacts when 'local' is enabled ($CLOUD_FILESYSTEM_PATH).
      --s3.endpoint=S3.ENDPOINT  The endpoint to use to connect to a Amazon S3 compatible cloud storage provider ($AWS_ENDPOINT).
      --s3.accessKeyId=S3.ACCESSKEYID
                                 The Amazon S3 Access Key Id ($AWS_ACCESS_KEY_ID).
      --s3.secretKey=S3.SECRETKEY
                                 The Amazon S3 secret key ($AWS_SECRET_ACCESS_KEY).
      --s3.region=S3.REGION      The Amazon S3 region($AWS_S3_REGION_NAME).
      --version                  Show application version.
```
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

Token can also be specified as enviroment variable `TURBO_TOKEN=xxxxx`, for example: `TURBO_TOKEN=xxxx turbo run build`

NOTE: CLI argument `--team` is not supported, please only use `teamId` from `.turbo/config.json`.

The `teamId` in `.turbo/config.json` is
used to generate a bucket in the cloud storage provider, as the id might be an
invalid name the team identifier a MD5 hash is generated and used as the bucket name.

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
