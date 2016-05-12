FROM golang

# Copy the runtime dockerfile into the context as Dockerfile
COPY Dockerfile.run /go/bin/Dockerfile

COPY esmap.go /go/src/github.com/mantika/esmap/esmap.go

WORKDIR /go/src/github.com/mantika/esmap

RUN go get -v -d ./...

RUN CGO_ENABLED=0 go build -a -installsuffix nocgo -o /go/bin/esmap .

# Set the workdir to be /go/bin which is where the binaries are built
WORKDIR /go/bin

# Export the WORKDIR as a tar stream
CMD tar -cf - .
