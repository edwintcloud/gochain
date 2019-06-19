#!/bin/bash  
ADDR1=$(go run main.go createwallet 2>&1)
ADDR2=$(go run main.go createwallet 2>&1)
ADDR1=${ADDR1:16}
ADDR2=${ADDR2:16}
echo -e "$(tput setaf 2)\nCreated Wallet $(tput setaf 6) $ADDR1"
echo -e "$(tput setaf 2)Created Wallet $(tput setaf 6) $ADDR2\n"
echo -e "$(tput setaf 2)Creating blockchain with Wallet $(tput setaf 6) $ADDR1 $(tput sgr0)\n"
go run main.go create -address $ADDR1
echo -e "$(tput setaf 2)\nPrint current blockchain$(tput sgr0)\n"
go run main.go print
echo -e "$(tput setaf 2)\nSend 30 tokens from $(tput setaf 6) $ADDR1 $(tput setaf 2) to $(tput setaf 6) $ADDR2 $(tput sgr0)\n"
go run main.go send -from $ADDR1 -to $ADDR2 -amount 30
echo -e "$(tput setaf 2)\nPrint blockchain again$(tput sgr0)\n"
go run main.go print
echo -e "$(tput setaf 2)\nGet balance of Wallet $(tput setaf 6) $ADDR1 $(tput sgr0)\n"
go run main.go getbal -address $ADDR1
echo -e "$(tput setaf 2)\nGet balance of Wallet $(tput setaf 6) $ADDR2 $(tput sgr0)\n"
go run main.go getbal -address $ADDR2
rm -rf ./data