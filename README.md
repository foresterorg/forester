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

* Direct ISO boot only through EFI HTTP and BIOS iPXE sanboot
* Support for generic (netboot) images (https://odcs.fedoraproject.org/composes/production/latest-Fedora-ELN/compose/BaseOS/x86_64/os/images/boot.iso)
* Support for ostree/bootc via generic (netboot) images
* Uploading via URL
* Bootstrapping unknown hosts does not work (make discovery interactive?)
* Update documentation on the recent changes (template generation, note that iPXE will not work with SecureBoot)
* Create events table and store installation milestones (boot, ks, finish) and rendered templates in the database
* Change log level to debug for "finished request" log for range requests (blocks are 4096, 8192, 32768, 65536 or) for ISO HTTP EFI Boot workflow: `msg="finished request" method=GET path=/img/1/image.iso duration_ms=0s status=206 bytes=131072 trace_id=pBI45d1z`
* Detect installation IP address (shim + %pre curl + event table) and secure the default sshpw password with "ssh" CLI fully working
* Squash migrations and refactor table names to singular
* Perform power operation in a goroutine (simple scheduler)
* Improve hardcoded power cycle delay (configurable?)
* Implement pykickstart checking of kickstart content (generated template and ks)
* Importing shim signatures in discovery mode: https://lukas.zapletalovi.com/posts/2021/rhelcentos-8-shim-kernel-signatures/
* Ability to create/edit/show system comment
* Make SlogDualWriter optional (this is only useful for debugging)
