# Configuration

Cloudpost project is configured using a `cloudpost.yml` file in the project root. This file defines the components that need to be deployed, as well as the configuration for each of the stages.

## Configuration file

The configuration file is a simple single-document YAML file called `cloudpost.yml` in the root of the project. A simple example can be found in [the example folder](/example/cloudpost.yml). It contains two keys:

- `components`, a list of all the components of the flock; and
- `config` representing the per-stage configuration of the flock.

## Component definitions

Each component must have `type` and `name` fields. The remaining fields depend on the component type. Names must be unique.

### Functions (`type` = `function`)

Creates a serverless function.

- `entry`: the entry-point function. This function must be defined in the `main` module. E.g. if the `entry` is `test_function`, sending the request to the service will invoke `main.test_function(requestObject)`.
- `triggerTopic`: the topic that the function will be implicitly subscribed to.
- `files`: the list of files in *filespec* format that should be deployed together.

Example:

```yml
- type: function
  name: example_function
  entry: example # will call main.example(...)
  files:
    - ['src/example_function/*', '.'] # copies everything to root
    - 'lib/library.py'                # copies to `lib/library.py`
  triggerTopic: topic_name # `topic_name` will trigger the function
```


### Container (`type` = `container`)

Creates a storage container.

- `entry`: the entry-point module of the container. The module should either export a WSGI application `app`, or a function `create_app()` that returns one.
- `triggerTopic`: the topic that the container will be implicitly subscribed to.
- `triggerPath`: the path that the implicitly subscribed topic will send the requests to.
- `files`: the list of files in *filespec* format that should be deployed together.
Example:

```yml
- type: container
  name: example_container
  entry: example # will serve `example.app`
  files:
    - ['src/example_container/*', '.'] # copies everything to root
  triggerTopic: topic_name # `topic_name` will send to `/hello`
  triggerPath: /hello
```

### Bucket (`type` = `bucket`)

Creates a storage bucket.

## Filespec format

Cloudpost uses *filespec* for specifying which files should be included in each deployment, and where they should be placed in the deployed artifact. A filespec is a list of entries. Each entry is one of the following:

- `path/to/file`: the file will be copied with the same path
- `path/with/glob/*`: all files will be copied with the same path
- `['path/to/source', 'path/to/destination']`: the file `path/to/source` will be copied to the `path/to/destination`. If the source is a directory, it will be copied entirely. Multiple directories can be copied to the same destination, in which case they will be merged. E.g.
  ```yml
  - ['./a', './dst']
  - ['./b', './dst']
  ```
  will create a `dst` directory in the output, containing the merged contents of `a` and `b`.
- `['path/so/glob/*', 'path/to/destination']`: all files matching the glob will be copied to the `path/to/destination`.


## Configuration

Configuration is described in terms of "layers". Each key in the `config` dictionary of the configuration file defines a layer. The base layer is always called `default`, and all the other layers are named after their stages. The following configuration

```yml
config:
  default:
    key_1: value_1
    key_2: value_2
  prod:
    key_1: value_1_override
  local:
    key_3: value_3
```

defines the `prod` and `local` environments equivalent to the following:

```yml
prod:
  key_1: value_1_override
  key_2: value_2
local:
  key_1: value_1
  key_2: value_2
  key_3: value_3
```