#!/usr/bin/env bash

package=""  # default to empty package
target=""   # efault to empty target

# parse options for this script's commands
while getopts ":h" opt; do
  case ${opt} in
    h )
      echo "usage:"
      echo "    ${0} -h                 display this help message"
      echo "    ${0} test <package>     run any tests for <package>"
      exit 0
      ;;
   \? )
     echo "invalid option: -$OPTARG" 1>&2
     exit 1
     ;;
  esac
done
shift $((OPTIND -1))

subcommand=$1; shift  # remove script name from the argument list
case "$subcommand" in
  # parse options for the test sub command
  test)
    package=$1; shift  # remove 'test' from the argument list

    # process package options
    while getopts ":t:" opt; do
      case ${opt} in
        t )
          target=$OPTARG
          ;;
        \? )
          echo "invalid option: -$OPTARG" 1>&2
          exit 1
          ;;
        : )
          echo "invalid option: -$OPTARG requires an argument" 1>&2
          exit 1
          ;;
      esac
    done
    shift $((OPTIND -1))
    ;;
esac