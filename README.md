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

* Refactor extracting with xorriso and prepare RPM/ostree docs
* Ability to pass whole kickstart via --ks option suppressing any templating
* Implement installation "queue" as a table (install uuid, exclusive, system id, creation time, state: queued, shim, grub, kickstart, done)
* Move image from system to installation table (maybe others)
* Set the global bootstrap shim image id from latest queue installation and drop global configuration
* Implement a scheduler for installation (when queue changes, a notification fires up on insert/update, calls go, it picks up the work)
* Queue processor never schedules more than one exclusive installation and checks in regular interval for new work (in case something get stucked)
* Support for loading shim via MAC address through Redfish boot URL param (/bmac/AA:BB:CC:DD:EE:FF/shim.efi)
* Importing shim signatures in dicovery mode: https://lukas.zapletalovi.com/posts/2021/rhelcentos-8-shim-kernel-signatures/
* Detect installation IP address (shim + %pre curl) and secure the default sshpw password with "ssh" CLI fully working
* Implement pykickstart checking of kickstart content (generated template and ks)
* Investigate how much work is BIOS support (via dnsmasq)
