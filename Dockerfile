FROM golang AS build
WORKDIR /root/
COPY ./src/ .
RUN CGO_ENABLED=0 GO111MODULE=on go build

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=build /root/pws-idokep-dispatcher /usr/bin
CMD ["/usr/bin/pws-idokep-dispatcher"]
# ENTRYPOINT [ "/bin/sh" ]