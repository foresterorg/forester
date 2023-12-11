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

* Bootstrapping unknown hosts does not work
* Restarting BIOS hosts via Redfish gives "value UefiHttp for the property BootSourceOverrideTarget is not in the list of acceptable values"
* Update documentation on the recent changes
* Create events table and store installation milestones (boot, ks, finish) and rendered templates in the database
* Change log level to debug for "finished request" log for range requests (blocks are 4096, 8192, 32768, 65536 or): `msg="finished request" method=GET path=/img/1/image.iso duration_ms=0s status=206 bytes=131072 trace_id=pBI45d1z`
* Detect installation IP address (shim + %pre curl + event table) and secure the default sshpw password with "ssh" CLI fully working
* Squash migrations and refactor table names to singular
* Implement a scheduler for power operations (when queue changes, a notification fires up on insert/update, calls go, it picks up the work)
* When image id is the same, use it. When there are different images add a warning message with sleep 1 minute.
* Scheduler will never start more than one system with different image id and checks in regular interval for new work
* Investigate how much work is BIOS support (via dnsmasq)
* Implement pykickstart checking of kickstart content (generated template and ks)
* Importing shim signatures in discovery mode: https://lukas.zapletalovi.com/posts/2021/rhelcentos-8-shim-kernel-signatures/
* Ability to create/edit/show system comment¨
* Make SlogDualWriter optional (this is only useful for debugging)
* Figure out if HTTPBootURL will work or HttpBootUrl is needed: https://github.com/stmcginnis/gofish/pull/271
