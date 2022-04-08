#!/bin/bash

cd error_codes
make

echo "-----Successful Termination-----"
./client http://localhost:5044 work
echo "----Unsuccessful Termination----"
./client http://localhost:5044 broken
echo "-----------------------------"

make clean
