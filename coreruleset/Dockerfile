FROM alpine:latest

RUN apk add git curl
RUN rm -rf /coraza/*
WORKDIR /coraza/coreruleset
RUN git clone https://github.com/coreruleset/coreruleset /coraza/coreruleset
RUN curl https://raw.githubusercontent.com/corazawaf/coraza/v2/master/coraza.conf-recommended -o /coraza/coraza.conf
# we replace SecRuleEngine DetectionOnly by SecRuleEngine On on /coraza/coraza.conf
RUN sed -i 's/SecRuleEngine DetectionOnly/SecRuleEngine On/g' /coraza/coraza.conf
RUN sed -i 's/SecDefaultAction "phase:1,log,auditlog,pass"/SecDefaultAction "phase:1,log,auditlog,deny,status:403"/g' /coraza/coreruleset/crs-setup.conf.example
RUN sed -i 's/SecDefaultAction "phase:2,log,auditlog,pass"/SecDefaultAction "phase:2,log,auditlog,deny,status:403"/g' /coraza/coreruleset/crs-setup.conf.example

