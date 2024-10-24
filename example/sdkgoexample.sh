#!/usr/bin/env bash
#
# argparser: true
#
# App Configuration:
# arg:-n,--name NAME:type=str:Application name
# opt:-d,--debug:type=flag:Enable debug mode
# opt:-p,--port PORT:type=int:Server port number
# opt:-t,--timeout VALUE:type=float:Operation timeout in seconds
# opt:-l,--log-file FILE:type=str:Optional log file path
# opt:-v,--verbose:type=flag:Enable verbose logging
# opt:-w,--workers COUNT:type=int:Number of worker threads
# opt:--throttle VALUE:type=float:Throttle rate
# opt:-H,--host VALUE:type=array/str:Server endpoints in host:port format
# opt:--features FEATURE:type=array/bool:Feature flags
# opt:--weights WEIGHT:type=array/float:Weight values
#
# Database Configuration:
# arg:--db-host HOST:type=str:Database host address
# arg:--db-port PORT:type=int:Database port
# arg:--db-user USER:type=str:Database username
# opt:--db-pass PASS:type=str:Database password
# opt:--db-replica VALUE:type=array/str:Database replicas in host:port format
#
# help: Example script demonstrating omnicli argument parsing
# +:
# +: This shows the usage of required arguments, optional flags,
# +: single values and arrays, with various supported types.

# Determine script dir
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Run the Go example
go run "${DIR}"/main.go "$@"
