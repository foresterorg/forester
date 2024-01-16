#!/bin/sh
set -ex
#
# This script is meant for DEVELOPMENT ONLY. Do not run this if you intend to use Forester.
#
go build -o forester-cli ./cmd/cli
URL=${URL:-http://localhost:8000}

./forester-cli --url "$URL" appliance create -n noop -k noop -u noop:///
./forester-cli --url "$URL" appliance create -n libvirt-system -k libvirt -u unix:///var/run/libvirt/libvirt-sock
./forester-cli --url "$URL" appliance create -n libvirt-session -k libvirt -u qemu:///session

./forester-cli --url "$URL" system register -n discovery -m 00:00:00:00:00:00 -f discovery=yes -a noop -u discovery
./forester-cli --url "$URL" system register -n dummy1 -m aa:bb:cc:dd:ee:f1 -f dummy=yes -a noop -u uid1
./forester-cli --url "$URL" system register -n dummy2 -m aa:bb:cc:dd:ee:f2 -f dummy=yes -a noop -u uid2

./forester-cli --url "$URL" image upload -n dummy-netboot fixtures/iso/fixture-netboot.iso
./forester-cli --url "$URL" image upload -n dummy-liveimg fixtures/iso/fixture-liveimg.iso

cat <<EOS | ./forester-cli --url "$URL" snippet create -i -n DiscoveryPre -k pre
# Use this snippet to deploy host with hardware address 00:00:00:00:00:00
# in order to discover boot unknown systems. Do not remove the following line:
{{ template "ks_discover.tmpl.py" . }}

# You may add additional %pre sections in order to perform additional commands
# during discovery. Example:
#%pre
#wipefs -a /dev/sd*
#%end
EOS

cat <<EOS | ./forester-cli --url "$URL" snippet create -i -n FedoraRPM39 -k source
url --mirrorlist="https://mirrors.fedoraproject.org/mirrorlist?repo=fedora-39&arch=x86_64"
repo --name=fedora-updates --mirrorlist="https://mirrors.fedoraproject.org/mirrorlist?repo=updates-released-f39&arch=x86_64"

%packages
@core
%end
EOS

cat <<EOS | ./forester-cli --url "$URL" snippet create -i -n SingleVolumeDisk -k disk
zerombr
clearpart --all
part /boot/efi --asprimary --fstype=vfat --label EFI --size=200
part / --asprimary --size=1024 --grow
bootloader --timeout=1
EOS

cat <<EOS | ./forester-cli --url "$URL" snippet create -i -n InsecureRootPw -k rootpw
# root/root
rootpw --iscrypted $6$.VxgzAA.37fSpLhE$KN4Cr92uYFYdndautiEd6jjd3p.C.0lzevLx5lGSPj5s.UHGocp57RG5bF2/HgFKmCW.fQd1gF0qn8et6i6WY/
EOS

cat <<EOS | ./forester-cli --url "$URL" snippet create -i -n NoSecurity -k security
firstboot --disabled
firewall --disabled
selinux --disabled
sshpw --username lzap --sshkey ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEhnn80ZywmjeBFFOGm+cm+5HUwm62qTVnjKlOdYFLHN lzap
EOS

cat <<EOS | ./forester-cli --url "$URL" snippet create -i -n HelloPost -k post
%post
echo "Hello Anaconda"; logger "Hello Forester"
%end
EOS

./forester-cli --url "$URL" system deploy 00:00:00:00:00:00 -i dummy-netboot -s DiscoveryPre -d "99999h"
curl "$URL/ks"

./forester-cli --url "$URL" system deploy aa:bb:cc:dd:ee:f1 -i dummy-liveimg -s HelloPost

curl -s "$URL/boot/shim.efi" >/dev/null
curl -s "$URL/boot/grubx64.efi" >/dev/null
curl "$URL/boot/mac/aa:bb:cc:dd:ee:f1"
curl -H "X-RHN-Provisioning-MAC-1: eth0 aa:bb:cc:dd:ee:f1" "$URL/ks"

./forester-cli --url "$URL" system deploy aa:bb:cc:dd:ee:f2 -i dummy-netboot -s FedoraRPM39 -s NoSecurity
./forester-cli --url "$URL" system release aa:bb:cc:dd:ee:f2

