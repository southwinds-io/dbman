#
#   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
#   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
#   Contributors to this project, hereby assign copyright in this code to the project,
#   to be licensed under the same terms as the rest of the code.
#
# NOTE: see https://access.redhat.com/solutions/4643601
# dnf checks for subsciptions even if not needed (not using RHEL repos)

FROM registry.access.redhat.com/ubi8/ubi-minimal
LABEL author="southwinds.io"
LABEL maintainer="admin@southwinds.io"
ARG UNAME=dbman
ENV UID=1000
ENV GID=1000
RUN microdnf install shadow-utils.x86_64 && \
    groupadd -g $GID -o $UNAME && \
    useradd -M -u $UID -g $GID $UNAME && \
    rm -rf /var/cache/yum && \
    microdnf clean all
WORKDIR /app
# copy dbman binaries
COPY dbman dbman-db-* /app/
# copy config files to user home
# 1. if running in podman or docker user home is "/home/dbman/"
COPY .dbman.toml .dbman_default.toml /home/dbman/
# 2. if running in openshift user home is "/"
COPY .dbman.toml .dbman_default.toml /
USER $UNAME
CMD ["/app/dbman", "serve"]
EXPOSE 8085/tcp
