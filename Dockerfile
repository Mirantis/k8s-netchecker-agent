FROM busybox

MAINTAINER Artem Roma <aroma@mirantis.com>

COPY _output/agent /usr/bin/netchecker-agent

ENTRYPOINT ["netchecker-agent", "-logtostderr"]

CMD ["-v=5"]
