FROM golang AS build
WORKDIR /root/
COPY ./src/ .
RUN CGO_ENABLED=0 GO111MODULE=on go build

FROM debian:bullseye-slim
COPY --from=build /root/dispatcher /usr/bin
CMD ["/usr/bin/dispatcher"]
# ENTRYPOINT [ "/bin/sh" ]