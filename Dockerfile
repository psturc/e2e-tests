FROM registry.ci.openshift.org/openshift/release:golang-1.20 AS builder

WORKDIR /github.com/redhat-appstudio/e2e-tests
USER root

COPY go.mod .
COPY go.sum .
RUN go mod download -x
COPY cmd/ cmd/
COPY magefiles/ magefiles/
COPY pkg/ pkg/
COPY tests/ tests/
COPY Makefile .

RUN go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
RUN ginkgo build ./cmd


FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

WORKDIR /root/
COPY --from=builder /go/bin/ginkgo /usr/local/bin
COPY --from=builder /github.com/redhat-appstudio/e2e-tests/cmd/cmd.test .
