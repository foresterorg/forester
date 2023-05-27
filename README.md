**Forester Project**

Bare-metal provisioning service for Red Hat Anaconda (Fedora, RHEL, CentOS Stream, Alma Linux...)

Requirements:

* Go 1.20
* Postgres

Hacking

Build the project:

    git clone https://github.com/foresterorg/forester
    cd forester
    go mod download
    ./build.sh

Create postgres database and execute all SQL files from internal/db/tern manually (migrations are work in progress).

Check possible environmental variables:

    ./forester-controller -h

Start the backend controller:

    DATABASE_NAME=mydb IMAGED_DIRECTORY=/my/images ./forester-controller

Start using the CLI:

    ./forester-cli --help

**License**

GNU GPL v3

Copyright (c) 2022 Lukáš Zapletal and AUTHORS, (c) 2023 Red Hat, Inc.

**TODO**

* https://github.com/webrpc/webrpc
* Out of band management
* Inventory