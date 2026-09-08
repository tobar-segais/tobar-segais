# The image is the binary and nothing else.
#
# Tobar Segais needs no runtime: no JVM, no application server, no database.
# The 1.x releases needed all three, and the image that carried them was the
# larger part of what an operator had to trust and to patch.

FROM --platform=$BUILDPLATFORM golang:1.26 AS build
WORKDIR /src

# The module files first, so a change to the code does not fetch the
# dependencies again.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# BUILDPLATFORM above and TARGETOS/TARGETARCH here: the compiler runs on the
# machine doing the building and produces a binary for the machine the image is
# for, which is what makes a multi-architecture build cost one compile per
# architecture rather than one emulated build per architecture.
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=""
RUN CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" \
    go build -trimpath \
      -ldflags="-s -w -X github.com/tobar-segais/tobar-segais/internal/version.version=${VERSION}" \
      -o /out/tobar-segais ./cmd/tobar-segais

# A directory to serve, so the image runs before anything is mounted over it.
# It cannot be made in the final stage: scratch holds no shell to make it with.
RUN mkdir -p /out/content

FROM scratch
COPY --from=build /out/tobar-segais /tobar-segais
COPY --from=build --chown=65532:65532 /out/content /content

# A number and not a name: scratch holds no /etc/passwd for a name to be in.
USER 65532:65532
EXPOSE 8080
VOLUME ["/content"]

ENTRYPOINT ["/tobar-segais"]
CMD ["serve", "--content", "/content", "--addr", ":8080"]
