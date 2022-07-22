ARG GOVERSION=1.17

FROM golang:${GOVERSION}-alpine AS build

WORKDIR /go/src/waf

RUN apk add --no-cache git libc-dev bash sed
RUN wget https://raw.githubusercontent.com/jptosso/coraza-waf/v2/master/coraza.conf-recommended
RUN git clone https://github.com/coreruleset/coreruleset
RUN sed -i 's/SecRuleEngine DetectionOnly/SecRuleEngine On/g' coraza.conf-recommended
RUN sed -i 's/SecDefaultAction "phase:1,log,auditlog,pass"/SecDefaultAction "phase:1,log,auditlog,deny,status:403"/g' coreruleset/crs-setup.conf.example
RUN sed -i 's/SecDefaultAction "phase:2,log,auditlog,pass"/SecDefaultAction "phase:2,log,auditlog,deny,status:403"/g' coreruleset/crs-setup.conf.example
COPY . .

RUN go install && go build -o waf

FROM alpine:latest
RUN apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community bash
RUN mkdir /src
WORKDIR /src
COPY --from=build /go/src/waf/waf /src/waf
COPY --from=build /go/src/waf/coraza.conf-recommended /src/coraza.conf
COPY --from=build /go/src/waf/coreruleset /src/coreruleset
ENTRYPOINT [ "/src/waf" ]
