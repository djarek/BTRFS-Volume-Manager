#!/bin/bash
export CGO_CFLAGS="-I"$(pwd)"/libbtrfs"
export CGO_LDFLAGS="-L"$(pwd)"/libbtrfs/build"


