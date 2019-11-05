Name: gpu-manager
Version: %{version}
Release: %{commit}%{?dist}
Summary: GPU Manager Plugin for Kubernetes

License: MIT
Source: gpu-manager-source.tar.gz

Requires: systemd-units

%define pkgname %{name}-%{version}-%{release}

%description
GPU Manager Plugin for Kubernetes

%prep
%setup -n gpu-manager-%{version}


%build
make all

%install
install -d $RPM_BUILD_ROOT/%{_bindir}
install -d $RPM_BUILD_ROOT/%{_unitdir}
install -d $RPM_BUILD_ROOT/etc/gpu-manager

install -p -m 755 ./go/bin/gpu-manager $RPM_BUILD_ROOT/%{_bindir}/
install -p -m 755 ./go/bin/gpu-client $RPM_BUILD_ROOT/%{_bindir}/

install -p -m 644 ./build/extra-config.json $RPM_BUILD_ROOT/etc/gpu-manager/
install -p -m 644 ./build/gpu-manager.conf $RPM_BUILD_ROOT/etc/gpu-manager/
install -p -m 644 ./build/volume.conf $RPM_BUILD_ROOT/etc/gpu-manager/

install -p -m 644 ./build/gpu-manager.service $RPM_BUILD_ROOT/%{_unitdir}/

%clean
rm -rf $RPM_BUILD_ROOT

%files
%config(noreplace,missingok) /etc/gpu-manager/extra-config.json
%config(noreplace,missingok) /etc/gpu-manager/gpu-manager.conf
%config(noreplace,missingok) /etc/gpu-manager/volume.conf

/%{_bindir}/gpu-manager
/%{_bindir}/gpu-client

/%{_unitdir}/gpu-manager.service
