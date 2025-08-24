## Certy

Certy provides an easy way to manage X.509 certificates and their private keys through a web interface. Using Certy, you have the ability to:

- Create root Certificate Authorities
- View certificate attributes
- Create intermediate certificates signed by the root CAs
- Export certificates and keys in PEM and PKCS#12 formats

### Screenshots

Here are some images of Certy in action:

<img src="https://github.com/nathan-osman/certy/blob/main/dist/ex-home.png?raw=true" width="250" /> &nbsp; <img src="https://github.com/nathan-osman/certy/blob/main/dist/ex-new.png?raw=true" width="250" /> &nbsp; <img src="https://github.com/nathan-osman/certy/blob/main/dist/ex-view.png?raw=true" width="250" />

### Building

To build the application, simply run:

    go build

### Installation

To install the application on Linux, you can use the handy "install" subcommand:

    sudo ./certy install

On Windows, you can use an elevated command prompt to run:

    .\certy.exe install
    .\certy.exe start

This will install Certy as a Windows Service.

> **Note:** running Certy on Windows is possible but not recommended since file & folder permissions are not yet correctly set during certificate creation. This will eventually be fixed but is a security issue in the meantime. Linux is not affected by this.
