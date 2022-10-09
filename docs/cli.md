# Using the Cloudpost CLI

## Using LocalRunner

LocalRunner is a local emulation of the cloud deployment. It can be used for development and debugging of the cloudpost flock.

To start the development server, run `cloudpost run`. This will start up all the services, and the local emulation of the cloud stack. This will give you the following APIs.

- For each function `function_name`:
    - `POST /<function_name>/event`: send a new request to a cloud function
    - `POST /<function_name>/pubsub`: send a new pubsub message to the function (automatically encodes and formats the message).
- For each container `container_name`:
    - `* /<container_name>/**`: send a request to the container. The path is container name is removed from the path before forwarding the request.

Aditionally, a web interface is hosted on `/ui/`.