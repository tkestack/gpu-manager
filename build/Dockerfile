ARG base_img
FROM nvidia/cuda:10.1-devel-centos7 as build

ARG version
ARG commit

RUN yum install -y rpm-build make

# default git has problems while cloning some repository
RUN yum install -y https://repo.ius.io/ius-release-el7.rpm \
  && yum install -y git222

ENV GOLANG_VERSION 1.14.3
RUN curl -sSL https://dl.google.com/go/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p /root/rpmbuild/{SPECS,SOURCES}

COPY gpu-manager.spec /root/rpmbuild/SPECS
COPY gpu-manager-source.tar.gz /root/rpmbuild/SOURCES

RUN echo '%_topdir /root/rpmbuild' > /root/.rpmmacros \
  && echo '%__os_install_post %{nil}' >> /root/.rpmmacros \
  && echo '%debug_package %{nil}' >> /root/.rpmmacros
WORKDIR /root/rpmbuild/SPECS
RUN rpmbuild -bb --quiet \
  --define 'version '${version}'' \
  --define 'commit '${commit}'' \
  gpu-manager.spec

FROM $base_img

ARG version
ARG commit

COPY --from=build /root/rpmbuild/RPMS/x86_64/gpu-manager-${version}-${commit}.el7.x86_64.rpm /tmp

RUN yum install epel-release -y && \
  yum install -y which jq

# Install packages
RUN rpm -ivh /tmp/gpu-manager-${version}-${commit}.el7.x86_64.rpm \
	&& rm -rf /tmp/gpu-manager-${version}-${commit}.el7.x86_64.rpm

# kubelet
VOLUME ["/var/lib/kubelet/device-plugins"]

# gpu manager storage
VOLUME ["/etc/gpu-manager/vm"]
VOLUME ["/etc/gpu-manager/vdriver"]
VOLUME ["/var/log/gpu-manager"]

# nvidia library search location
VOLUME ["/usr/local/host"]

RUN echo "/usr/local/nvidia/lib" > /etc/ld.so.conf.d/nvidia.conf && \
    echo "/usr/local/nvidia/lib64" >> /etc/ld.so.conf.d/nvidia.conf

ENV PATH=$PATH:/usr/local/nvidia/bin

# cgroup
VOLUME ["/sys/fs/cgroup"]

# display
EXPOSE 5678

COPY start.sh /
COPY copy-bin-lib.sh /

CMD ["/start.sh"]
