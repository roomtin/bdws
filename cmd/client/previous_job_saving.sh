#!/bin/bash

gcc example.c -o example

echo "-----------Original Run------------"
./client http://localhost:5044 example
echo "---Replay of Previous Job Result---"
cat ../supervisor/lastjob.txt
echo "-----------------------------------"

rm -f example
