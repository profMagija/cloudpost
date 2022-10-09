# Configuration

Cloudpost project is configured using a `cloudpost.yml` file in the project root. This file defines the components that need to be deployed, as well as the configuration for each of the stages.

## Component definitions

Each component must have `type` and `name` fields. The remaining fields depend on the component type. Names must be unique.

### Functions (`type` = `function`)

Creates a serverless function.

- `entry`: the entry-point function. This function must be defined in the `main` module. E.g. if the `entry` is `test_function`, sending the request to the service will invoke `main.test_function(requestObject)`.
- `triggerTopic`: the topic that the function will be implicitly subscribed to.
- `files`: the list of files in *filespec* format that should be deployed together.


### Container (`type` = `container`)

Creates a storage container.

- `entry`: the entry-point module of the container. The module should either export a WSGI application `app`, or a function `create_app()` that returns one.
- `triggerTopic`: the topic that the container will be implicitly subscribed to.
- `triggerPath`: the path that the implicitly subscribed topic will send the requests to.
- `files`: the list of files in *filespec* format that should be deployed together.

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
