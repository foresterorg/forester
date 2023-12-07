#!/bin/bash
set -xe
SRCDIR="{{ .ImageDir }}"
DSTDIR="{{ .ImageDir }}"

TROOT=$(mktemp -d /tmp/forester-troot-XXXXXXX)
TAUX=$(mktemp -d /tmp/forester-taux-XXXXXXX)
trap "rm -rf $TROOT $TAUX" EXIT
rm -f $DSTDIR/boot.iso

if [[ -d $SRCDIR/boot/grub2/i386-pc ]]; then
  PCGRUBDIR=$SRCDIR/boot/grub2/i386-pc
else
  PCGRUBDIR=/usr/lib/grub/i386-pc
fi

mkdir -p $TROOT/images/pxeboot $TROOT/boot/grub2/i386-pc $TROOT/EFI/BOOT
cp $SRCDIR/images/pxeboot/* $TROOT/images/pxeboot
cp $SRCDIR/EFI/BOOT/grub.cfg $TROOT/EFI/BOOT
cp $PCGRUBDIR/* $TROOT/boot/grub2/i386-pc
(test -f $SRCDIR/LICENSE && cp $SRCDIR/LICENSE $TROOT) || true

cat >$TROOT/boot/grub2/grub.cfg <<EOBIOS
# Generated by FORESTER (BIOS) version {{ .Version }}
function load_video {
  insmod all_video
}
load_video
set gfxpayload=keep
insmod gzio
insmod part_gpt
insmod ext2
insmod chain
search --no-floppy --set=root -l 'FORESTER'
linux /images/pxeboot/vmlinuz inst.stage2={{ .BaseURL }}/img/{{ .ImageID }} ip=dhcp inst.text inst.sshd inst.ks.sendmac inst.ks={{ .BaseURL }}/ks inst.syslog={{ .BaseHost }}:{{ .SyslogPort }} systemd.hostname=f-${net_default_mac}
initrd /images/pxeboot/initrd.img
boot
EOBIOS

cat >$TROOT/EFI/BOOT/grub.cfg <<EOEFI
# Generated by FORESTER (EFI) version {{ .Version }}
function load_video {
  insmod efi_gop
  insmod efi_uga
  insmod video_bochs
  insmod video_cirrus
  insmod all_video
}
load_video
set gfxpayload=keep
insmod gzio
insmod part_gpt
insmod ext2
search --no-floppy --set=root -l 'FORESTER'
linuxefi /images/pxeboot/vmlinuz inst.stage2={{ .BaseURL }}/img/{{ .ImageID }} ip=dhcp inst.text inst.sshd inst.ks.sendmac inst.ks={{ .BaseURL }}/ks inst.syslog={{ .BaseHost }}:{{ .SyslogPort }} systemd.hostname=f-${net_default_mac}
initrdefi /images/pxeboot/initrd.img
boot
EOEFI

truncate -s 8M $TAUX/efiboot.img
mkfs.vfat $TAUX/efiboot.img
mmd -i $TAUX/efiboot.img ::/EFI ::/EFI/BOOT ::/EFI/BOOT/fonts
mcopy -i $TAUX/efiboot.img $SRCDIR/EFI/BOOT/BOOTX64.EFI ::EFI/BOOT
mcopy -i $TAUX/efiboot.img $SRCDIR/EFI/BOOT/grubx64.efi ::EFI/BOOT
mcopy -i $TAUX/efiboot.img $SRCDIR/EFI/BOOT/fonts/unicode.pf2 ::EFI/BOOT/fonts
mcopy -i $TAUX/efiboot.img $TROOT/EFI/BOOT/grub.cfg ::EFI/BOOT

grub2-mkimage -O i386-pc-eltorito -d $PCGRUBDIR \
  -o $TROOT/images/eltorito.img \
  -p /boot/grub2 \
  iso9660 biosdisk

# Create hybrid (xorriso from 2022 required) and if it fails create EFI-only
xorrisofs -o $DSTDIR/boot.iso \
  -R -J -V 'FORESTER' \
  --grub2-mbr $PCGRUBDIR/boot_hybrid.img \
  -partition_offset 16 \
  -appended_part_as_gpt \
  -append_partition 2 C12A7328-F81F-11D2-BA4B-00A0C93EC93B $TAUX/efiboot.img \
  -iso_mbr_part_type EBD0A0A2-B9E5-4433-87C0-68B6B72699C7 \
  -c boot.cat --boot-catalog-hide \
  -b images/eltorito.img \
  -no-emul-boot -boot-load-size 4 -boot-info-table --grub2-boot-info \
  -eltorito-alt-boot \
  -e '--interval:appended_partition_2:all::' -no-emul-boot \
  -graft-points \
  $TROOT || \
xorrisofs -o $DSTDIR/boot.iso \
  -R -J -V 'FORESTER' \
  -isohybrid-mbr /usr/share/syslinux/isohdpfx.bin \
  -boot-load-size 4 -boot-info-table -no-emul-boot \
  -eltorito-alt-boot -e images/efiboot.img -no-emul-boot \
  -b isolinux.bin -c boot.cat \
  -graft-points \
  $TROOT \
  images/efiboot.img=$TAUX/efiboot.img \
  isolinux.bin=/usr/share/syslinux/isolinux.bin

grub2-mkimage -O i386-pc-pxe -o $DSTDIR/grubx64.0 -p / tftp pxe normal ls echo minicmd halt reboot http linux

exit 0