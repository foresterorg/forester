**Forester Project**

Bare-metal image-based unattended provisioning service for Red Hat Anaconda
(Fedora, RHEL, CentOS Stream, Alma Linux...) which works out-of-box. It
utilises Redfish API and UEFI HTTP Boot to deploy images created by Image
Builder through Anaconda.

More information, quick start and documentation available at
https://foresterorg.github.io

**API**

The service API is RPC over HTTP with [OpenAPI Specification](https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/foresterorg/forester/main/openapi.gen.yaml)

**Feedback and support**

Visit our [discussion forums](https://github.com/foresterorg/forester/discussions)!

**License**

GNU GPL v3

Copyright (c) 2022 Lukáš Zapletal and AUTHORS, (c) 2023 Red Hat, Inc.

**TODO**

* Snippets / templates for customization of partitioning and post scripts via table SNIPPET(ID,TYPE,BODY)
* Default snippets are hardcoded (if-else template statement)
* Refactor: move both kind types into client-shared package and remove kindToInt/intToKind
* API/CLI for logs
* Support for hash in `liveimg --checksum=<sha256>`: SHA of ISO itself and image tarball via table CHECKSUM(ID,OBJECT,ALG,SUM)
* Validate kickstart of existing host (CLI -> calls ksvalidate)
* BIOS support (at least investigate how much work is this)
