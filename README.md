# Shecan Diagnostic

This tool gathers network diagnostics related to the [Shecan](https://shecan.ir) DNS service. It checks DNS connectivity, pings Shecan servers, performs lookups and sends a diagnostic report.

## Building

```bash
# build a binary named `shecan-diagnostic`
go build -o shecan-diagnostic
```

For reproducible cross-platform builds you can use
[Task](https://taskfile.dev). This will generate binaries for Linux, macOS and
Windows under the `build/` directory:

```bash
# install Task if not already available
go install github.com/go-task/task/v3/cmd/task@latest

# build cross-platform binaries under the `build/` directory
task build-all
```

## Usage

Run the binary directly to start the diagnostic workflow. You can optionally supply a plan flag (`--plan` or `-p`) with values `Free` or `Pro`.

```bash
# run with interactive plan selection
./shecan-diagnostic

# or explicitly set the plan
./shecan-diagnostic --plan Free
```

The command `run` is the default action and executes automatically when no arguments are provided.

## Docker

You can also build and run the application using Docker:

```bash
# build the Docker image
docker build -t shecan-diagnostic .

# run the container
docker run --rm -it shecan-diagnostic
```

Set `REPORT_SERVER_URL` in the environment if you want to send
results to your own server:

```bash
docker run --rm -it -e REPORT_SERVER_URL=https://example.com/report shecan-diagnostic
```
