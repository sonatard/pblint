# pblint

pblint is a command line tool to lint Protocol Buffers.

## Install

```console
go get -u github.com/sonatard/pblint
```

## Usage

```console
pblint -i proto/ api/v1/*.proto
```

## Lint Rules

1. File name must be `servicename_service.proto`.
1. Request type must be in rpc declared file.
1. Response type must be in rpc declared file.
1. Request type name must be `MethodNameRequest`.
1. Response type name must be `MethodNameResponse`.
1. HTTP rule must set.
1. HTTP Method use GET or POST.
1. HTTP URL must use `/ServiceName/MethodName`.
1. HTTP Body must be `*`.
1. HTTP Body must not use AdditionalBindings.
1. Other Message must not be in `servicename_service.proto`.

