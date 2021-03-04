FROM debian:10

RUN apt update && \
  apt-get install -y curl git && \
  curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 && \
  chmod 700 get_helm.sh && \
 ./get_helm.sh

RUN apt update && \
  apt install -y golang-go

RUN curl -fsSL https://get.docker.com -o get-docker.sh && \
    chmod 700 get-docker.sh && \
   ./get-docker.sh
