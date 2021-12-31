# README

In this example, we will be using the open-source solution Minio, which offers an API that is compatible with Amazon S3. In our case, the data of the bucket will be stored on the file system. The example shows how you can connect to a cloud provider that is compatible with the Amazon S3 API endpoints. 

Typically, when you are planning to use S3 you won't need to define this as the server will use the defaults for Amazon S3 when it's not given. But it illustrates how you could use an Amazon S3 compatible cloud storage provider, such as Walabi, Blackblaze or Scaleway

This example, the main addition is using the `AWS_ENDPOINT` to specify the location of the API endpoint. 

*This is for demonstration purposes only*

## Getting Started

You can run the Tapico Turborepo Remote cache and Minio example by running the
following command:

```bash
cd examples/with-docker-compose
docker-compose up
```

The  docker-compose will do the following things:

* Start an instance Minio, an Amazon S3 compatible storage
* Create a bucket named `tapico-remote-cache`
* Build a docker image and start the Tapico Turborepo Remote cache service

The service gets preconfigured with the following token: `2a25157f-3ff9-440a-b549-b0fe7fb4a5ea`
that can be used when using Turbo. The cache artefacts send to the service will
be stored in the `data`-directory.

If you want to change the list of allowed Turbo tokens, you can change the `TURBO_TOKEN`
environment variable in the docker-compose file. The expected format of the value is
a comma separated list of strings. You could consider to add the token that Vercel
generated when you linked your Vercel account with Turbo. You should be able to
look up the token on the Vercel Console web-site.

## Configuring Turbo

Turbo needs to be configured so that it won't be using the Vercel remote cache but
instead our locally running alternative remote-cache. If you use the docker compose
file mentioned in the previous section, you will have the remote cache running on
`http://127.0.0.1:8080`.

If so, you can either configure Turbo by changing the `.turbo/config.json` configuration
file or by passing arguments when executing the `run`-command.

### Configuring the `config.json`-file

You can update the `.turbo/config.json` file by updating the `apiUrl`-property
to point it to our local remote cache api server.

```json
{
  "teamId": "team_blah",
  "teamSlug": "turborepo-monorepo-experiment",
  "apiUrl": "http://127.0.0.1:8080"
}
```

The `teamSlug`-property is a unique name that can be used to give a more human-readable
name to the project.  The `teamId`-property is the identifier for the team. You can either
provide both or only `teamId`.

*Note*: If the remote cache server is not running with the argument `--enable-bucket-per-team`
and the `teamSlug`-property is provided then the server will use the value of this property
to create a subdirectory in the bucket and store the cache artefacts.

You can now run the turbo repo with the following command to use the remote cache
while executing the `build` run script:

```bash
turbo run build --token=2a25157f-3ff9-440a-b549-b0fe7fb4a5ea
```

### Use arguments

If you don't want to change the `.turbo/config.json` file, you can also pass the appropriate
arguments to the Turbo CLI. You would need to pass the `--token`, `--api` arguments.

For example, you can use the command below to execute a run script named `build`:

```bash
turbo run test --api="http://127.0.0.1:8080" --token=2a25157f-3ff9-440a-b549-b0fe7fb4a5ea -vvv
```

## Exploring bucket

### Using Minio Console

If you want to explore which cache artefacts have been uploaded to local Minio instance,
you can login into the Minio console by navigating to: [http://127.0.0.1:9001](Minio Consle)
and use the credentials:

* Username: `minio`
* Password: `miniosecretkey`

After you have logged in you, select [Buckets](127.0.0.1:9001/buckets) from the
navigation bar on the left-hand side, the page will show all the available Buckets
and how much storage has been used.

A bucket named `tapico-remote-cache` should be listed on this page and you can
use `Browse`-button to explore its contents.

### Use the file system

This example stores the contents of the bucket on the host's file system in the
subdirectory named `data` inside `examples/with-docker-compose`. This directory
should have a directory named `tapico-remote-cache` which contains all the
uploaded cache artefacts.

## After thoughts

This example has been created to allow you to conveniently try out the service and
is for demonstration purposes. It's not advised to use this Docker compose file in
production.

You might want to use a production-ready solution such as Amazon S3
and AWS Fargate, or Google Cloud Storage in combination with Cloud Run. Both allow
you to run the service in a cost-effective manner.

*Warning*: The docker compose script will delete the contents of the bucket
when it gets restarted.
