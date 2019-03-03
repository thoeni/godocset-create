#!/bin/bash
export GOPATH=/go

while getopts ":t:o:f:" opt; do
  case $opt in
    t) arg_token="$OPTARG"
    ;;
    o) arg_org="$OPTARG"
    ;;
    f) arg_filter="$OPTARG"
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    ;;
  esac
done

rm -rf /tmp/godoc-$arg_org.docset

echo "godoc-create -githubToken $arg_token -organization $arg_org && godocdash -name godoc-$arg_org -silent -filter $arg_filter && mv godoc-$arg_org.docset /tmp"
godoc-create -githubToken $arg_token -organization $arg_org -filter $arg_filter && godocdash -name godoc-$arg_org -silent -filter $arg_filter && mv godoc-$arg_org.docset /tmp