## Table of contents

- [Development](#development)
- [DB Migration](#db-migration)

## Development

This section describes the steps to have the deviceregistry backend up and running on your local dev setup.

### App Configuration

Make sure to create a copy of the configuration template files that will be used by your local backend
instance:

```bash
$ cp config.tpl.yml config-development.yml
```

The naming convention for the config files here is important, because this is the way the configuration
management library identifies to correct file to be parsed.

#### Environmental variables
The application also supports configuration parameters to be injected as envs.

The pattern that it expects should follow the following rules and it relies on the same structure define on `config.tpl.yml`:
- All envs should start with prefix `DEVICEREGISTRY`
- All letters to be uppercase
- Nested elements will be linked by `_`
- Separation characters of `-` should be replaced by `_`

Some examples:

| yaml    | env                |
|---------|--------------------|
| foo     | DEVICEREGISTRY_FOO     |
| foo.bar | DEVICEREGISTRY_FOO_BAR |
| foo-bar | DEVICEREGISTRY_FOO_BAR |

There are some `make` helpers to facilitate your Docker related commands:

#### Spin up services

```shell
$ make docker-up
```

#### List services

List the newly created services to make sure everything is up and running:

```shell
$ make docker-ps
```

#### Manual image build

```shell
$ make build
```

#### Update the backend service

If you want to update the image artifact with your recent code changes and spin it up at the same time:

```shell
$ make docker-update
```

## DB Migration

- `go run main.go migrate help` to display migration help
- `go run main.go migrate up` to migrate up
- `go run main.go migrate up-by-one` to migrate up by one step
- `go run main.go migrate down` to migrate down
- `go run main.go migrate status` to display migration status
- `go run main.go --env=test migrate` to migrate in another environment (here: `test`)

