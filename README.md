# Mousiki

[![](https://github.com/nlowe/mousiki/workflows/CI/badge.svg)](https://github.com/nlowe/mousiki/actions)  [![License](https://img.shields.io/badge/license-MIT-brightgreen)](./LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/nlowe/mousiki)](https://goreportcard.com/report/github.com/nlowe/mousiki)

A command-line pandora client. Inspired by [PromyLOPh/pianobar](https://github.com/PromyLOPh/pianobar).

## Usage

TODO

## Building

You need Go 1.12+ or vgo for Go Modules support. You also need a C compiler that works with CGO and gstreamer-1.0-devel

```bash
# Linux / Unix / macOS
go build -o mousiki main.go

# Windows
go build -o mousiki.exe main.go
```

## License

This plugin is published under the MIT License. See [`./LICENSE`](./LICENSE) for details.
