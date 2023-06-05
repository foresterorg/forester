**Forester Project**

Bare-metal provisioning service for Red Hat Anaconda (Fedora, RHEL, CentOS Stream, Alma Linux...)

**Requirements**:

* Go 1.20+
* Postgres

**Feedback and support**

Visit our [discussion forums](https://github.com/foresterorg/forester/discussions)!

**Hacking**

Build the project, the script will also install required CLI tools for code generation and database migration:

    git clone https://github.com/foresterorg/forester
    cd forester
    ./build.sh

Create postgres database, configure the migrator and run it:

    cat config/tern.conf
    [database]
    host = localhost
    port = 5432
    database = forester
    user = postgres

    ./migrate.sh

Check possible environmental variables:

    ./forester-controller -h

Start the backend controller:

    DATABASE_NAME=mydb IMAGED_DIRECTORY=/my/images ./forester-controller

Start using the CLI:

    ./forester-cli --help

**TLS**

- Sign server.cer with ca.cer
- Configure the service to use it (TBD)
- mkfs.msdos -C letsencrypt.img 300
- mcopy -i letsencrypt.img letsencrypt.cer ::/ca.cer
- mdir -i letsencrypt.img
- Enroll the certificate into EFI

**License**

GNU GPL v3

Copyright (c) 2022 Lukáš Zapletal and AUTHORS, (c) 2023 Red Hat, Inc.

**TODO**

* Out of band management
