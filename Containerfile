FROM quay.io/projectquay/golang:1.20 as build
USER 0
RUN mkdir /build
WORKDIR /build
COPY . .
RUN go build -o forester-controller cmd/controller/ctl_main.go

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
COPY --from=build /build/forester-controller /forester-controller
USER 1001
CMD ["/forester-controller"]
CMD microdnf install xorriso grub2-tools grub2-tools-extra syslinux syslinux-nonlinux dosfstools grub2-pc-modules pykickstart
