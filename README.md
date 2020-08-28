# Mousiki

[![](https://github.com/nlowe/mousiki/workflows/CI/badge.svg)](https://github.com/nlowe/mousiki/actions)  [![License](https://img.shields.io/badge/license-MIT-brightgreen)](./LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/nlowe/mousiki)](https://goreportcard.com/report/github.com/nlowe/mousiki)

A command-line pandora client. Inspired by [PromyLOPh/pianobar](https://github.com/PromyLOPh/pianobar).

![](doc/mousiki.png)

## Installation

Right now you have to build from source. See [Building](#building).

Mousiki relies on gstreamer-1.0 for audio playback. Install it from your package manager or from
https://gstreamer.freedesktop.org/download/.

## Usage

Start `mousiki` with the `--username` parameter, you will be prompted for
your password:

```bash
$ mousiki --username somebody@gmail.com
Password:
```

### Transport Controls

`mousiki` currently supports the following controls:

| Key | Action |
| --- | ------ |
| `N` | Skip to the next track (or fetch more tracks if the queue is empty |
| `E` | Explain why pandora picked the currently playing track |
| `<space>` | Pause / Resume playback |
| `+` | Love Song |
| `-` | Ban Song |
| `t` | Tired of Song |
| `esc` | Station Picker |
| `Q` / `Ctrl+C` | Quit |

## TODO

In no particular order:

* Publish binaries to GitHub
* Feedback Management (Thumbs Up/Down tracks) for recently played tracks
* Skip ads / artist messages automatically
* Audio quality selection / Support for premium qualities

Maybe some day:

* Pure-GO AAC Decoder to drop the CGO Dependency on gstreamer (or figure out how to get pandora to send us MP3 streams)
* OSC / HTTP API for controlling playback / running a playback server / writing custom frontends

## Building

You need Go 1.12+ or vgo for Go Modules support. You also need a C compiler that works with CGO and gstreamer-1.0-devel.

See [`.github/workflows/ci.yaml`](.github/workflows/ci.yaml) for a rough idea of what to install.

```bash
# Linux / Unix / macOS
go build -o mousiki main.go

# Windows
go build -o mousiki.exe main.go
```

## License

This plugin is published under the MIT License. See [`./LICENSE`](./LICENSE) for details.
