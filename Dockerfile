# The image must be built starting from a specific TFLint version.
# ARGs are cleared after each FROM so we need to declare this here.
ARG TFLINT_VERSION="v0.39.3"

# Use golang base image just to build our wrapper
FROM golang:1.19-alpine as builder

# Copy all the files, download dependneices and build
WORKDIR /usr/src/app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build

# ARGs are cleared after each FROM so we need to declare this here.
ARG TFLINT_VERSION

# This is the actual Code Climate analyzer
FROM ghcr.io/terraform-linters/tflint:${TFLINT_VERSION}

# ...so that you know who to blame if things don't work!
LABEL maintainer="Stefano Tenuta <stapps@outlook.it>"

# Code Climate's policy for Docker requires to run as an user with ID=9000
RUN adduser -u 9000 -D app

# Copy the already built wrapper from the Go builder image
COPY --from=builder /usr/src/app/codeclimate-tflint /usr/src/app/codeclimate-tflint

# ARGs are cleared after each FROM so we need to declare this here.
ARG TFLINT_VERSION

# Download rules for current version from GitHub, leveraging their subversion suppport
RUN apk add --no-cache subversion
WORKDIR /tflint-rules

# We don't want any subpath to be created so we --force to download in the current WORKDIR
RUN svn export https://github.com/terraform-linters/tflint/tags/${TFLINT_VERSION}/docs/rules ./ --force

# Svn is not needed anymore so we can delete it
RUN apk del subversion

# Final set up to follow Code Climate's Docker policy
RUN chown -R app:app /usr/src/app
USER app
VOLUME /code
WORKDIR /code

# We don't have to have an entrypoint but we must set our wrapper as CMD
ENTRYPOINT [""]
CMD ["/usr/src/app/codeclimate-tflint"]
