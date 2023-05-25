# Go HTTP Server with Customizable Configuration

This is a simple Go HTTP server that serves static content and JSON files. You can customize the server's behavior using environment variables, such as setting the server port, specifying multiple static content directories, and adding a delay to API requests.

## Features

- Serve static content from one or multiple directories
- Serve JSON data based on the HTTP method and URL path
- Customize server port, static content directories, and API request delay using environment variables

## Requirements

- Go 1.16 or higher

## Installation

1. Clone the repository:

```sh
git clone https://github.com/yksnyh/simple-json-server.git
cd simple-json-server
```

2. Build the binary:

```sh
go build -o simple-json-server server.go
```

## Usage

1. Start the server with default settings:

```sh
./simple-json-server
```

2. Customize the server using environment variables:

```sh
export SERVER_PORT=8080
export STATIC_CONTENT_DIRS="html,assets"
export API_REQUEST_DELAY_MS=1000
./simple-json-server
```

## Environment Variables

- `SERVER_PORT`: The port number for the server to listen on (default: `8888`).
- `STATIC_CONTENT_DIRS`: A comma-separated list of directories to serve static content from (default: `html`).
- `API_REQUEST_DELAY_MS`: The number of milliseconds to sleep before handling an API request (default: no delay).

## API Request Handling

The server looks for JSON files in the `data` directory based on the HTTP method and the URL path. For example, a `GET` request to `/api/users/1` would look for a file at `data/get/api/users/1.json`. If the file is found, its content is returned as the JSON response. If the file is not found, a 404 Not Found status with an error message is returned.

## License

This project is released under the MIT License.