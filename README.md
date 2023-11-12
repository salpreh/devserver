# devserver
Go cli to start a development server

## Quickstart
Currently, there are 2 main commands: `echo` and `mock` (use `--help` flag for more details).
- With `devserver echo` you can start a server that will echo the request body and headers.
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
          "message": "Que fas?"
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