# Certy

[![Build Status](https://github.com/nathan-osman/certy/actions/workflows/test.yml/badge.svg)](https://github.com/nathan-osman/certy/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/nathan-osman/certy/badge.svg?branch=main)](https://coveralls.io/github/nathan-osman/certy?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/nathan-osman/certy.svg)](https://pkg.go.dev/github.com/nathan-osman/certy)
[![MIT License](https://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](https://opensource.org/licenses/MIT)

Certy provides an easy way to manage X.509 certificates and their private keys through a web interface. Using Certy, you have the ability to:

- Create root Certificate Authorities
- View certificate attributes
- Create intermediate certificates signed by the root CAs
- Export certificates and keys in PEM and PKCS#12 formats
- Validate certificates and certificate chains
- Do all of this with a choice of light or dark theme!

### Screenshots

Here are some images of Certy in action:

<img src="https://github.com/nathan-osman/certy/blob/main/dist/ex-home.png?raw=true" width="250" /> &nbsp; <img src="https://github.com/nathan-osman/certy/blob/main/dist/ex-new.png?raw=true" width="250" /> &nbsp; <img src="https://github.com/nathan-osman/certy/blob/main/dist/ex-view.png?raw=true" width="250" />

### Installation

To download the application, visit the [releases page](https://github.com/nathan-osman/certy/releases/) and select the file that matches your platform.

To run the application on Linux with systemd, use the "install" and "start" subcommands:

    sudo ./certy install
    sudo ./certy start

On Windows, use an elevated command prompt to run:

    .\certy.exe install
    .\certy.exe start

This will install and start Certy as a Windows Service.

> **Note:** running Certy on Windows is possible but not recommended since file & folder permissions are not yet correctly set during certificate creation. This will eventually be fixed but is a security issue in the meantime. Linux is not affected by this.

### Docker

In addition to running as a standalone service, Certy can run in a Docker container. The command for launching Certy in Docker looks something like this:

    docker run \
        -d \
        --name certy \
        -p 8000:8000 \
        "$(pwd)/data:/data" \
        nathanosman/certy

This will launch the service listening on port 8000 on your host and store data in `data/` in the current directory.

### Building

To build the application, simply run:

    go build