# devserver
Go cli to start a development server

## Quickstart
Currently, there are 2 main commands: `echo` and `mock` (use `--help` flag for more details).
- With `devserver echo` you can start a server that will echo the request body, headers, method and path. Body is echoed in the response body, the rest of data is sent in headers.
- With `devserver mock` you can start a server that will mock a response based on a json file.

### Mock server
The file format for the mock server
```json lines
{
  "headers": { // Common headers
    "content-type": "application/json",
    "x-origin": "Mockserver"
  },
  "paths": { // Paths
    "/hello/boy": {
      "responses": { // Responses
        "200": {
          "content": "Hi there!",
          "name": "boy"
        },
        "500": {
          "code": "ER01",
          "message": "Do not touch!"
        }
      }
    },
    "/hello": {
      "headers": { // Headers for this path (overrides common headers)
        "content-type": "text/plain"
      },
      "responses": {
        "200": "World!"
      }
    },
    "/hello/name": {
      "get": { // Per method responses
        "responses": {
          "200": {
            "content": "Hi there!",
            "name": "name"
          },
          "400": {
            "code": "ER01",
            "message": "What is?"
          }
        }
      },
      "post": {
        "responses": {
          "204": null // No body content
        }
      }
    }
  }
}
```

## Installation
You will need to have [Go](https://golang.org/doc/install) installed in your machine to build the project.
With go available you can use the `install` Makefile target to install the binary in your `$GOBIN` folder.
```bash
make install
```