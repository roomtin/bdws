#!/bin/bash

gcc work.c -o work
gcc broken.c -o broken

echo "-----Successful Termination-----"
./client http://localhost:5044 work
echo "----Unsuccessful Termination----"
./client http://localhost:5044 broken
echo "-----------------------------"

rm -f work broken
