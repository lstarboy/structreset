#!/usr/bin/env bash

do_structreset=$(which structreset)
if [[ ${do_structreset}  == "" ]]; then
  echo "请先安装structreset"
fi

check_dir(){
  echo "check dir: " "$1"
  local has_go_file=false
  for file in "$1"/*
  do
    if test -f "$file"
    then
      if [[ $file == *.go ]]; then
        has_go_file=true
      fi
    fi

    if test -d "$file"
    then
        check_dir "$file"
    fi
  done

  if [[ ${has_go_file} == true ]]; then
    ${do_structreset} "$1"
  fi
}

print_help() {
    echo "Usage: structreset [FLAG] DIR/Package/Go File"
    echo -e "\nFLAG:"
    echo -e "  -d 需要检查的目录. eg: structreset -d ./logic "
    echo -e "  -h print help "
    echo -e ""
    echo -e "用法1：递归检查目录下所有包 structreset -d ./"
    echo -e "用法2：检查单个包 structreset ./logic/oldprotocol"
    echo -e "用法3：检查单个Go文件 structreset ./logic/oldprotocol/struct.go"
    echo -e ""
}

main(){
  if [[ $1 == "help" || $1 == "-h"  || $1 == "" ]]; then
    print_help
    exit 0
  fi

  if [[ $1 == -* ]]; then
    if [[ $1 == "-d" && $2 != "" ]]; then
      check_dir "$2"
      echo "finished check."
    else
      print_help
    fi
  else
    ${do_structreset} "$@"
    echo "finished check."
  fi
}

main "$@"

