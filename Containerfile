FROM registry.fedoraproject.org/fedora-minimal:latest as build
RUN microdnf install -y golang
RUN mkdir /build
WORKDIR /build
# ignore other (big) files like ISO images
COPY cmd internal go.mod go.sum .
RUN go build -o forester-controller cmd/controller/ctl_main.go

FROM registry.fedoraproject.org/fedora-minimal:latest
COPY --from=build /build/forester-controller /forester-controller
RUN microdnf install -y xorriso grub2-tools grub2-tools-extra syslinux syslinux-nonlinux dosfstools grub2-pc-modules pykickstart ipxe-bootimgs
RUN microdnf clean all
CMD ["/forester-controller"]
