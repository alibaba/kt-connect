#!/bin/sh
a=0
while true
do
   echo $a > /dev/null
   a=`expr $a + 1`
   sleep 2
done