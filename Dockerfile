FROM busybox

MAINTAINER Artem Roma <aroma@mirantis.com>

ADD _output/agent /opt/bin/agent

ENV PATH=$PATH:/opt/bin

ENTRYPOINT ["agent"]

CMD ["-v=5", "-alsologtostderr=true"]
