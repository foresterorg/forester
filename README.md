**Forester Project**

Bare-metal image-based unattended provisioning service for Red Hat Anaconda (Fedora, RHEL, CentOS Stream, Alma Linux...) which works out-of-box. It utilizes Redfish API and UEFI HTTP Boot to deploy images created by Image Builder through Anaconda.

**DevConf 2023 talk**: https://www.youtube.com/live/6nRP0si2wKI?feature=share&t=8674

**DEMO**: https://www.youtube.com/watch?v=jxCHU_nzluY

**WARNING**:

This project is currently in proof of concept, it is fully functional for PoC testing and [feedback](https://github.com/foresterorg/forester/discussions).

**Install via Podman**

The following commands create two new volumes for postgres database and images, a new pod named `forester` with two containers named `forester-pg` and `forester-api` exposing port 8000 for the REST API.

    podman volume create forester-pg
    podman volume create forester-img
    podman pod create --name forester -p 8000:8000
    podman run -d --name forester-pg --pod forester -e POSTGRESQL_USER=forester -e POSTGRESQL_PASSWORD=forester -e POSTGRESQL_DATABASE=forester -v forester-pg:/var/lib/pgsql/data:Z quay.io/fedora/postgresql-15; sleep 5s
    podman run -d --name forester-api --pod forester -e DATABASE_USER=forester -e DATABASE_PASSWORD=forester -e IMAGES_DIR=/img -v forester-img:/img:Z quay.io/forester/controller:latest

To uninstall everything up including data and images:

    podman rm -f forester-pg forester-app
    podman pod rm -f forester
    podman volume rm forester-img forester-pg

**Install via Docker Compose**

To start the Forester controller and Postgres database on port 8000 with data for both database and images stored in `./data` directory run:

    curl https://raw.githubusercontent.com/foresterorg/forester/main/compose.yaml > compose.yaml
    EXPOSED_APP_PORT=8000 DATA_DIR=./data podman-compose up -d

This command requires [podman-compose](https://github.com/containers/podman-compose) (available in Fedora and EPEL). You can use docker-compose as well, it is known to work but be aware we do not test against Docker during development.

To remove Forester completely, stop containers and remove them, and the data directory.

**Building images**

Build and download RHEL image from [console.redhat.com](https://console.redhat.com/insights/image-builder) or build your own Fedora or CentOS image using [osbuild](https://www.osbuild.org/) (also known as Red Hat Image Builder) or Lorax (legacy image builder):

    livemedia-creator --make-iso \
        --iso=Fedora-Server-dvd-x86_64-37-1.7.iso \
        --ks /usr/share/doc/lorax/fedora-minimal.ks \
        --image-name=f37-minimal-image.iso

**Redfish hardware setup**:

Servers need to boot via UEFI HTTP Boot (not UEFI PXE) a particular URL `http://forester:8000/boot/shim.efi` (where `forester` is a machine running the Forester controller container). There are currently two options how to achieve that.

First option, which is great for PoC or testing out Forester on just a handful of machines, is to configure the HTTP UEFI Boot URL in BIOS directly. For example, on DELL iDrac go to Configuration - BIOS Settings - Network Settings and enable HTTP device and set the URI. Then apply and commit the change (requires reboot). To speed the boot process up, disable PXE devices on the same page.

Second option is to configure the URL on the DHCP server so no changes are required through out of band management. Example configuration for ISC DHCPv4 (also works on DHCPv6):

```
class "httpclients" {
  match if substring (option vendor-class-identifier, 0, 10) = "HTTPClient";
  option vendor-class-identifier "HTTPClient";
  filename "http://forester:8000/boot/shim.efi";
}
subnet 192.168.42.0 netmask 255.255.255.0 {
  range dynamic-bootp 192.168.42.100 192.168.42.120;
  default-lease-time 14400;
  max-lease-time 172800;
}
```

Warning: only HTTP scheme is currently supported by the project.

**Configuring the service**

Download and extract the [CLI for your architecture](https://github.com/foresterorg/forester/releases) and try to connect to controller:

    ./forester-cli --url http://forester:8000 appliance list

If the service is running on localhost port 8000, you can omit the `--url` argument.

The first step is to upload the image:

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

Create appliance, for hacking and development a good appliance type is libvirt through local UNIX socket:

    ./forester-cli appliance create --kind 1 --name libvirt

To create a Redfish appliance use kind number 2:

    ./forester-cli appliance create --kind 2 --name dellr350 --uri https://root:calvin@dr350-a14.local

Warning: username and password are currently stored as clear text and fully readable through the API.

    ./forester-cli appliance list
    ID  Name     Kind  URI
    1   libvirt  1     unix:///var/run/libvirt/libvirt-sock
    2   dellr350 2     https://root:calvin@dr350-a14.local

Discover the system or multiple blades in chassis:

    ./forester-cli appliance enlist dellr350

One or more systems are available now, each system has an unique ID, one or more MAC addresses and randomly generated name. A system can be referenced via both MAC address and random name:

```
./forester-cli system list
ID  Name        Hw Addresses           Acquired  Facts
1   Lynn Viers  6c:fe:54:70:60:10 (4)  true      Dell Inc. PowerEdge R350
```

To show more details of a system:

```
./forester-cli system show Viers
Attribute       Value
ID              1
Name            Lynn Viers
Acquired        true
Acquired at     Mon Sep  4 14:40:50 2023
Image ID        1
MAC             6c:fe:54:70:60:10
MAC             c4:5a:b1:a0:f2:b5
MAC             6c:fe:54:70:60:11
MAC             c4:5a:b1:a0:f2:b4
Appliance Name  dell
Appliance Kind  2
Appliance URI   https://root:calvin@dell-r350-08-drac.mgmt.sat.rdu2.redhat.com
UID             4c4c4544-004c-3510-804c-c4c04f435731

Fact                     Value
baseboard-asset-tag      
baseboard-manufacturer   Dell Inc.
baseboard-product-name   0MTYYT
baseboard-serial-number  .DL5XXXX.MXWSG0000000HE.
baseboard-version        A02
bios-release-date        11/14/2022
bios-revision            1.5
bios-vendor              Dell Inc.
bios-version             1.5.1
chassis-asset-tag        Not Specified
chassis-manufacturer     Dell Inc.
chassis-serial-number    DL5XXXX
chassis-type             Rack Mount Chassis
chassis-version          Not Specified
cpuinfo-processor-count  4
firmware-revision        
memory-bytes             8201367552
processor-family         Xeon
processor-frequency      2800 MHz
processor-manufacturer   Intel
processor-version        Intel(R) Xeon(R) E-2314 CPU @ 2.80GHz
redfish_asset_tag        
redfish_description      Computer System which represents a machine (physical or virtual) and the local resources such as memory, cpu and other devices that can be accessed from that machine.
redfish_manufacturer     Dell Inc.
redfish_memory_bytes     8589934592
redfish_model            PowerEdge R350
redfish_name             System
redfish_oid              /redfish/v1/Systems/System.Embedded.1
redfish_part_number      0MTYYTA02
redfish_pcie_dev_count   9
redfish_processor_cores  4
redfish_processor_count  1
redfish_processor_model  Intel(R) Xeon(R) E-2314 CPU @ 2.80GHz
redfish_serial_number    MXWSJ0032100HI
redfish_sku              DL5XXXX
serial                   DL5XXXX
system-family            PowerEdge
system-manufacturer      Dell Inc.
system-product-name      PowerEdge R350
system-serial-number     DL5XXXX
system-sku-number        SKU=0A94;ModelName=PowerEdge R350
system-uuid              4c4c4544-004c-3510-804c-c4c04f435731
system-version           Not Specified
```

Facts which start with `redfish` were recognized via Redfish API, other facts can be discovered by booting the system into Anaconda in a released state:

    ./forester-cli appliance bootnet lynn

**Deploy images**

Systems are either **released** or **acquired**. By acquisition, an operator performs installation of a specific image onto the hardware:

    ./forester-cli system acquire lynn --imagename RHEL9

To release a system and put it back to the pool of available systems:

    ./forester-cli system release lynn

Warning: There is no authentication or authorization in the API, anyone can acquire or release systems or even add new appliances.

**Libvirt setup**:

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

**Compiling from sources**

Build the project, the script will also install required CLI tools for code generation and database migration:

    git clone https://github.com/foresterorg/forester
    cd forester
    ./build.sh

When you start the backend for the first time, it will migrate database (create tables). By default, it connect to "localhost" database "forester" and user "postgres".

    ./forester-controller

**Redfish emulators**

There are several emulators available.

The official one unfortunatelly does not allow updates (reboot), you can still try it out:

    podman run --rm -p 5000:5000 dmtf/redfish-interface-emulator:latest

However, there is another one from DMTF, which does support updates:

    podman run --rm -p 5000:8000 dmtf/redfish-mockup-server:latest


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
