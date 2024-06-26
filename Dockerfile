FROM golang:1-1.22-bookworm as builder
# Set the working directory to /app
WORKDIR /nebula-http-gateway
# Copy the current directory contents into the container at /app
COPY . /nebula-http-gateway
 # Make port available to the world outside this container
ENV GOPROXY https://goproxy.cn
RUN go build

FROM golang:1-1.22-bookworm

WORKDIR /root
COPY --from=builder ./nebula-http-gateway .
COPY ./conf ./conf

EXPOSE 8080

ENTRYPOINT ["./nebula-http-gateway"]
