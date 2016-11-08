FROM busybox

MAINTAINER Artem Roma <aroma@mirantis.com>

ADD dist/agent /opt/bin/agent

ENV PATH=$PATH:/opt/bin

ENTRYPOINT ["agent"]

# let's log at INFO level by default
CMD ["-v=4", "-alsologtostderr=true"]
