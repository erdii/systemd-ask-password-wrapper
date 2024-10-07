Name:           systemd-ask-password-wrapper
Version:        0.0.1
Release:        1%{?dist}
Summary:        Forward systemd-ask-password requests to a wrapped command

License:        GPLv3
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang
BuildRequires:  systemd-rpm-macros

Provides:       %{name} = %{version}

%description
Forward systemd-ask-password requests to a wrapped command

%global debug_package %{nil}

%prep
%autosetup

%build
go build -v -o %{name} ./cmd/systemd-ask-password-wrapper

%install
install -Dpm 0755 %{name} %{buildroot}%{_bindir}/%{name}
install -Dpm 644 config/%{name}.service %{buildroot}%{_unitdir}/%{name}.service
install -Dpm 644 config/%{name}.path %{buildroot}%{_unitdir}/%{name}.path

%check
# go test should be here... :)

%post
%systemd_post %{name}.service
%systemd_post %{name}.path

%preun
%systemd_preun %{name}.service
%systemd_preun %{name}.path

%files
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%{_unitdir}/%{name}.path
%config(noreplace) %{_sysconfdir}/%{name}/config.json


%changelog
* Sun Oct 6 2024 Josh Gwosdz - 0.0.1-1
- First release%changelog
