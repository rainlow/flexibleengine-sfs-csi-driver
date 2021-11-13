FROM centos:7.6.1810
LABEL maintainers="Dafu SHI"
LABEL description="Huawei Cloud SFS CSI Plugin compatible with Flexible Engine"

RUN yum -y install nfs-utils && yum -y install epel-release && yum -y install jq && yum clean all

COPY sfs-csi-plugin /sfs-csi-plugin

ENTRYPOINT ["/sfs-csi-plugin"]
