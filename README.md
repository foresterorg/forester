**Forester Project**

Bare-metal image-based unattended provisioning service for Red Hat Anaconda
(Fedora, RHEL, CentOS Stream, Alma Linux...) which works out-of-box. It
utilises Redfish API and UEFI HTTP Boot to deploy images created by Image
Builder through Anaconda.

More information, quick start and documentation available at
https://foresterorg.github.io

**API**

The service API is RPC over HTTP with [OpenAPI Specification](https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/foresterorg/forester/main/openapi.gen.yaml)

Language clients:

* [Python](https://github.com/foresterorg/forester-client-python)

**Feedback and support**

Visit our [discussion forums](https://github.com/foresterorg/forester/discussions)!

**License**

GNU GPL v3

Copyright (c) 2022 Lukáš Zapletal and AUTHORS, (c) 2023 Red Hat, Inc.

**TODO**

* Change name of existing (enlisted) system
* Custom snippet for acquisition (simple string)
* Restarting running libvirt VM does not work (shutdown, edit domain, start needed)
* Support for hash in `liveimg --checksum=<sha256>`: SHA of ISO itself and image tarball via table CHECKSUM(ID,OBJECT,ALG,SUM)
* Validate kickstart of existing host (CLI -> calls ksvalidate)
* Investigate how much work is BIOS support
