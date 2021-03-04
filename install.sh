#!/usr/bin/env bash


main(){
  go build -o structreset main.go

  cp ./structreset /usr/local/bin/structreset
  cp ./cmd/structresetx.sh /usr/local/bin/structresetx
}

main "$@"