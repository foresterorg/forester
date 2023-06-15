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
    Environment variables:
    
      APP_PORT int
            HTTP port of the API service (default "8000")
      APP_HOSTNAME string
            hostname of the service (default "")
      APP_INSTALL_DURATION int64
            duration for which the service initiates provisioning after acquire (default "1h")
      DATABASE_HOST string
            main database hostname (default "localhost")
      DATABASE_PORT uint16
            main database port (default "5432")
      DATABASE_NAME string
            main database name (default "forester")
      DATABASE_USER string
            main database username (default "postgres")
      DATABASE_PASSWORD string
            main database password (default "")
      DATABASE_MIN_CONN int32
            connection pool minimum size (default "2")
      DATABASE_MAX_CONN int32
            connection pool maximum size (default "50")
      DATABASE_MAX_IDLE_TIME int64
            connection pool idle time (time interval syntax) (default "15m")
      DATABASE_MAX_LIFETIME int64
            connection pool total lifetime (time interval syntax) (default "2h")
      DATABASE_LOG_LEVEL string
            logging level of database logs (default "trace")
      LOGGING_LEVEL string
            logger level (debug, info, warn, error) (default "debug")
      IMAGES_DIR string
            absolute path to directory with images (default "images")
      IMAGES_BOOT_ID int
            boot shim/grub from image DB ID (default "1")

Start the backend controller:

    IMAGES_DIR=/var/lib/forester ./forester-controller

Download RHEL image from console.redhat.com or build your own Fedora or CentOS image using [osbuild](https://www.osbuild.org/) (Image Builder) or Lorax (legacy image builder):

    livemedia-creator --make-iso \
        --iso=Fedora-Server-dvd-x86_64-37-1.7.iso \
        --ks /usr/share/doc/lorax/fedora-minimal.ks \
        --image-name=f37-minimal-image.iso

Start using the CLI:

    ./forester-cli --help

Upload the image:

    ./forester-cli image upload --name Fedora37 f37-minimal-image.iso

Check it:

    ./forester-cli image list
    Image ID  Image Name
    1         RHEL9
    2         Fedora37

    ./forester-cli image show Fedora37
    Attribute  Value
    ID         2
    Name       Fedora37

Configure libvirt environment for booting from network via UEFI HTTP Boot, add the five "dnsmasq" options into the "default" libvirt network:

    sudo virsh net-edit default
    <network xmlns:dnsmasq='http://libvirt.org/schemas/network/dnsmasq/1.0'>
      <name>default</name>
      …
      <dnsmasq:options>
        <dnsmasq:option value='dhcp-vendorclass=set:efi-http,HTTPClient:Arch:00016'/>
        <dnsmasq:option value='dhcp-option-force=tag:efi-http,60,HTTPClient'/>
        <dnsmasq:option value='dhcp-boot=tag:efi-http,&quot;http://192.168.122.1:8000/boot/shim.efi&quot;'/>
      </dnsmasq:options>
    </network>

Make sure to update the HTTP address in case you want to use different network than "defalut" (which is 192.168.122.0). Restart the network to make the DHCP server settings effective:

    sudo virsh net-destroy default
    sudo virsh net-start default

Now, boot an empty UEFI VM on this network from network, it must be set for UEFI HTTP Boot. Enter EFI firmware by pressing ESC key after it is powered on, enter "Boot manager" screen and make "UEFI HTTPv4 [MAC]" the first boot option.

Boot the VM into discovery mode.

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

* System registration finds and updates existing systems
* Discovery CLI
* Remove "json" tags from IDL
