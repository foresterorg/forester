FROM quay.io/projectquay/golang:1.21 as build
RUN mkdir /build
WORKDIR /build
COPY . .
RUN go build -o forester-controller cmd/controller/ctl_main.go

FROM registry.fedoraproject.org/fedora-minimal:39
COPY --from=build /build/forester-controller /forester-controller
RUN microdnf -y install xorriso grub2-tools grub2-tools-extra syslinux syslinux-nonlinux dosfstools grub2-pc-modules pykickstart ipxe-bootimgs
CMD ["/forester-controller"]
