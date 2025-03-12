# erGO

erGO is a state machine–inspired workflow engine written in Go. It supports multiple operator types (Bash, CLI, Python, Map/Reduce) and error handling (with fallback commands), similar to Snakemake, Nextflow, GCP Workflows, and AWS Step Functions.

## Installation

Clone the repository and build the application:

```bash
go build -o ergo ./cmd/ergo


## ./ergo --run ErgoFile

