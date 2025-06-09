# Shecan Diagnostic

This tool gathers network diagnostics related to the [Shecan](https://shecan.ir) DNS service. It checks DNS connectivity, pings Shecan servers, performs lookups and sends a diagnostic report.

## Building

```bash
# build a binary named `shecan-diagnostic`
go build -o shecan-diagnostic
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
