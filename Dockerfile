FROM golang:1.22-alpine3.18 AS go-builder
#ARG arch=x86_64

# See https://github.com/CosmWasm/wasmvm/releases
ENV LIBWASMVM_VERSION=v2.0.0

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git cmake
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /code
COPY . /code/

# Install mimalloc
RUN git clone --depth 1 https://github.com/microsoft/mimalloc; cd mimalloc; mkdir build; cd build; cmake ..; make -j$(nproc); make install
ENV MIMALLOC_RESERVE_HUGE_OS_PAGES=4

# See https://github.com/\!cosm\!wasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/${LIBWASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${LIBWASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a

# Highly recommend to verify the version hash
# RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep a5e63292ec67f5bdefab51b42c3fbc3fa307c6aefeb6b409d971f1df909c3927
# RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep 762307147bf8f550bd5324b7f7c4f17ee20805ff93dc06cc073ffbd909438320
# Copy the library you want to the final location that will be found by the linker flag `-linitia_muslc`

RUN cp /lib/libwasmvm_muslc.`uname -m`.a /lib/libwasmvm_muslc.a

# force it to use static lib (from above) not standard libwasmvm.so file
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LDFLAGS="-linkmode=external -extldflags \"-L/code/mimalloc/build -lmimalloc -Wl,-z,muldefs -static\"" make build

FROM alpine:3.18

RUN addgroup minitia \
    && adduser -G minitia -D -h /minitia minitia

WORKDIR /minitia

COPY --from=go-builder /code/build/minitiad /usr/local/bin/minitiad

# for new-metric setup
COPY --from=go-builder /code/contrib /minitia/contrib

USER minitia

# rest server
EXPOSE 1317
# grpc
EXPOSE 9090
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

CMD ["/usr/local/bin/minitiad", "version"]
