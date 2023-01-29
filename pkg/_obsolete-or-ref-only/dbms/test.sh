#!/usr/bin/env bash

package=`pwd`

gobin=`which go`
echo "$gobin"

# shift our nargs count
shift 0 # right now we don't need to shift

# usage function
function usage {
  echo "usage function"
}

# check for usage
if [[ "$#" != 2 ]]; then
  usage # run usage function
fi

# page tests function
function page_tests {
  echo ":: running page tests..."
  cd page
  $gobin test -v -run '^TestPage.*$'
  cd -
  echo ":: page tests are done"
}

# check for page test case
if [[ "$1" == "test" && "$2" == "page" ]]; then
  page_tests # run page tests
fi

# disk tests function
function disk_tests {
  echo ":: running disk tests..."
  cd disk
  $gobin test -v -run '^TestDisk.*$'
  $gobin test -v -run '^TestSegment.*$'
  cd -
  echo ":: disk tests are done"
}

# check for disk test cases
if [[ "$1" == "test" && "$2" == "disk" ]]; then
  disk_tests
fi

# buffer pool tests function
function buffer_tests {
  echo ":: running buffer tests..."
  cd buffer
  $gobin test -v -run '^TestBuffer.*$'
  $gobin test -v -run '^TestClockReplacer.*$'
  cd -
  echo ":: buffer tests are done"
}

# check for buffer test cases
if [[ "$1" == "test" && "$2" == "buffer" ]]; then
  buffer_tests
fi

# all tests
function all_tests {
  page_tests
  disk_tests
  buffer_tests
}

# check for all test cases
if [[ "$1" == "test" && "$2" == "all" ]]; then
  all_tests
fi

#if %1%==pageTests call :runPageTests
#if %1%==diskTests call :runDiskTests
#if %1%==bufferTests call :runBufferTests
#
#:runPageTests
#"%gobin%" test dbms/page -run "^TestPage.*$"
#exit /b 0
#
#:runDiskTests
#"%gobin%" test dbms/disk -run "^TestDisk.*$"
#exit /b 0
#
#:runBufferTests
#"%gobin%" test dbms/buffer -run "^TestBuffer.*$"
#exit /b 0