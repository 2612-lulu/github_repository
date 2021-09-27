#!/bin/bash

echo "run client server of transaction!"
set NODE_NAME
export NODE_NAME="C1"
gnome-terminal -t "C1" -x bash -c "./qbrun.exe transaction -from 1CG9GcxF2BT1rjxwoMHSLtRP9RTCsSkRyH -to 1TJhYPkgZ7xngyu137D4VHvn2EySuwbEh -amount 2;exec bash"
