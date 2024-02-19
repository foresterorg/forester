FROM registry.fedoraproject.org/fedora-minimal:latest as build
RUN microdnf install -y golang
RUN mkdir /build
WORKDIR /build
COPY . .
RUN go build -o forester-controller ./cmd/controller
RUN go build -o forester-proxy ./cmd/proxy
RUN go build -o forester-cli ./cmd/cli

FROM registry.fedoraproject.org/fedora-minimal:latest
RUN microdnf install -y xorriso grub2-tools grub2-tools-extra syslinux syslinux-nonlinux dosfstools grub2-pc-modules pykickstart ipxe-bootimgs
RUN microdnf clean all
COPY --from=build /build/forester-controller /build/forester-proxy /build/forester-cli /
CMD ["/forester-controller"]
