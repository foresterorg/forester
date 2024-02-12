FROM quay.io/projectquay/golang:1.21 as build
USER 0
RUN mkdir /build
WORKDIR /build
COPY . .
RUN go build -o forester-controller cmd/controller/ctl_main.go

FROM registry.access.redhat.com/ubi9/ubi-init:latest
COPY --from=build /build/forester-controller /forester-controller
RUN subman register --username "<user>" --password "<pwd>"
RUN dnf -y install xorriso grub2-tools grub2-tools-extra syslinux syslinux-nonlinux dosfstools grub2-pc-modules pykickstart ipxe-bootimgs
USER 1001
CMD ["/forester-controller"]
