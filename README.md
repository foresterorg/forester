**Forester Project**

Bare-metal provisioning service for Red Hat Anaconda (Fedora, RHEL, CentOS Stream, Alma Linux...)

**DevConf 2023 talk**: https://www.youtube.com/live/6nRP0si2wKI?feature=share&t=8674

**DEMO**: https://www.youtube.com/watch?v=jxCHU_nzluY

**Requirements**:

* Go 1.20+
* Postgres

**Hardware setup**:

Configure your servers for UEFI HTTP Boot and set the UEFI HTTP Boot URL to `http://forester:8000/boot/shim.efi`. This can be either done via DHCP or in the BIOS. For example, on DELL iDrac go to Configuration - BIOS Settings - Network Settings and disable PXE devices but enable HTTP device and set the URI. Then apply and commit the change (requires reboot).

**Hacking**

Build the project, the script will also install required CLI tools for code generation and database migration:

    git clone https://github.com/foresterorg/forester
    cd forester
    ./build.sh

Create postgres database, configure the migrator and run it:

    go install github.com/jackc/tern/v2@latest
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

Create appliance, the only supported one is libvirt through local UNIX socket:

    ./forester-cli appliance create -name libvirt

Kind is automatically set to 1 (libvirt) for now:

    ./forester-cli appliance list
    ID  Name     Kind  URI
    1   libvirt  1     unix:///var/run/libvirt/libvirt-sock

Configure libvirt environment for booting from network via UEFI HTTP Boot, add the five "dnsmasq" options into the "default" libvirt network. Also, optionally configure PXEv4 and IPEv6 to return a non-existing file ("USE_HTTP" in the example) to speed up OVMF firmware to fallback to HTTPv4:

    sudo virsh net-edit default
    <network xmlns:dnsmasq='http://libvirt.org/schemas/network/dnsmasq/1.0'>
      <name>default</name>
      <uuid>9f3e4377-3d33-42df-b34c-7880295d24ee</uuid>
      <forward mode='nat'/>
      <bridge name='virbr0' zone='trusted' stp='on' delay='0'/>
      <mac address='52:54:00:7a:00:01'/>
      <ip address='192.168.122.1' netmask='255.255.255.0'>
        <tftp root='/var/lib/tftpboot'/>
        <dhcp>
          <range start='192.168.122.2' end='192.168.122.254'/>
          <bootp file='USE_HTTP'/>
        </dhcp>
      </ip>
      <ip family='ipv6' address='2001:db8:dead:beef:fe::2' prefix='96'>
        <dhcp>
          <range start='2001:db8:dead:beef:fe::1000' end='2001:db8:dead:beef:fe::2000'/>
        </dhcp>
      </ip>
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

**Redfish emulators**

There are several emulators available. The official one

    podman run --rm -p 5000:5000 dmtf/redfish-interface-emulator:latest

does not allow updates (reboot). However, there is another one from DMTF

    podman run --rm -p 5000:8000 dmtf/redfish-mockup-server:latest

which works.

**TLS**

- Sign server.cer with ca.cer
- Configure the service to use it (TBD)
- mkfs.msdos -C letsencrypt.img 300
- mcopy -i letsencrypt.img letsencrypt.cer ::/ca.cer
- mdir -i letsencrypt.img
- Enroll the certificate into EFI

**Feedback and support**

Visit our [discussion forums](https://github.com/foresterorg/forester/discussions)!

**License**

GNU GPL v3

Copyright (c) 2022 Lukáš Zapletal and AUTHORS, (c) 2023 Red Hat, Inc.

**TODO**

* Fix bugs
* After restart, new EFI entry is not booted by qemu and system boots from network again
* Cockpit has "Install" button instead of "Start" button for the initial installation (but still use `<boot dev='hd'/>`)
