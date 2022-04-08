#!/bin/bash

cd error_codes
make

echo "-----------Original Run------------"
./client http://localhost:5044 example
echo "---Replay of Previous Job Result---"
cat ../supervisor/lastjob.txt
echo "-----------------------------------"

make clean
