FROM debian:bullseye

RUN \
apt-get update && \
apt-get install -y \
apt-utils \
apt-transport-https \
build-essential \
cmake \
curl \
git \
gnupg2 \
iproute2 \
iputils-ping \
libboost-all-dev \
libeigen3-dev \
libglib2.0-dev \
libgmp-dev \
libntl-dev \
libssl-dev \
locales \
net-tools \
procps \
software-properties-common \
tcpdump \
time \
wget

RUN cd /tmp && \
git clone https://github.com/spring-epfl/emp-tool.git ; cd emp-tool ; cmake . ; make ; make install ; cd /tmp ; rm -r emp-tool && \
for k in ot zk sh2pc ag2pc agmpc; do git clone https://github.com/emp-toolkit/emp-${k}.git ; cd emp-${k}; cmake . ; make ; make install ; cd /tmp ; rm -r emp-${k} ; done && \
echo "EMP Installed!"

# UTF-8 support in console.
RUN \
printf '%s\n' 'fr_CH.UTF-8 UTF-8' 'en_US.UTF-8 UTF-8' >> /etc/locale.gen && \
printf '%s\n' 'LANG="en_US.UTF-8"' 'LANGUAGE="en_US:en"' >> /etc/default/locale && \
dpkg-reconfigure --frontend=noninteractive locales && \
update-locale 'LANG=en_US.UTF-8' && \
echo "export VISIBLE=now" >> /etc/profile

VOLUME  ["/psi"]
