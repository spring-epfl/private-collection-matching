FROM debian:bullseye
RUN echo 'deb http://deb.debian.org/debian bullseye-backports main' >> /etc/apt/sources.list && \
export DEBIAN_FRONTEND='noninteractive' && \
apt-get update && \
apt-get upgrade -y && \
apt-get install -y \
apt-transport-https \
ca-certificates \
curl \
ipython3 \
less \
nano \
python3-dev \
python3-pip \
python3-venv \
vim \
&& \
apt-get install -t bullseye-backports -y golang-go && \
apt-get clean -y

RUN python3 -m pip install rdkit seaborn

COPY ./ /GoPSI
RUN mkdir -p /GoPSI/chemistry
RUN mkdir -p /GoPSI/data
RUN cd /GoPSI && go get ./...
RUN cd /GoPSI/cmd/doc_search && go build
RUN cd /GoPSI/cmd/chem_search && go build
RUN cd /GoPSI/cmd/small_domain_bench && go build

# Chembl database location
VOLUME ["/GoPSI/chemistry"]
VOLUME ["/GoPSI/data"]

WORKDIR /GoPSI
