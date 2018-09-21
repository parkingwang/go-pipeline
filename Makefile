# Binary name
BINARY=go-pipeline
VERSION=1.1.1

GITTAG=`git rev-parse --short HEAD`
BUILD_TIME=`date +%FT%T%z`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.GitHash=${GITTAG} -X main.BuildTime=${BUILD_TIME} -X main.Version=${VERSION}"
BUFLAGS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Release
BUILD_DIR=./build

# Binary
BUILD_OUT_ORIGIN="${BUILD_DIR}/${BINARY}.origin.bin"
BUILD_OUT_ZIP="${BUILD_DIR}/${BINARY}.bin"

# Control
BUILD_CTRL_ORIGIN="${BUILD_DIR}/ctrl.origin.bin"
BUILD_CTRL_ZIP="${BUILD_DIR}/ctrl.bin"

# Builds the project
build:
		rm -rf ${BUILD_DIR}
		mkdir -p ${BUILD_DIR}

		# Build for linux
		${BUFLAGS} go build ${LDFLAGS} -o ${BUILD_OUT_ORIGIN} ./cmd/main.go
		${BUFLAGS} go build ${LDFLAGS} -o ${BUILD_CTRL_ORIGIN} ./cmd/ctrl.go

		# Compress
		upx -o ${BUILD_OUT_ZIP} ${BUILD_OUT_ORIGIN}
		rm ${BUILD_OUT_ORIGIN}

		upx -o ${BUILD_CTRL_ZIP} ${BUILD_CTRL_ORIGIN}
		rm ${BUILD_CTRL_ORIGIN}

		# Copy configs and scripts
		cp -R ./cmd/conf.d ${BUILD_DIR}
		cp ./cmd/*.sh ${BUILD_DIR}

		# Permissions
		chmod +x ${BUILD_DIR}/*.bin
		chmod +x ${BUILD_DIR}/*.sh

		# Write version
		echo "${VERSION}" > ${BUILD_DIR}/version

install:
		go install

clean:
		go clean

.PHONY:  clean build