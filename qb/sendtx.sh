#!/bin/bash

echo "run client server of transaction!"
set NODE_NAME
export NODE_NAME="C1"
gnome-terminal -t "C1" -x bash -c "./qbrun.exe transaction -from 1CG9GcxF2BT1rjxwoMHSLtRP9RTCsSkRyH -to 1TJhYPkgZ7xngyu137D4VHvn2EySuwbEh -amount 2;exec bash"

export NODE_NAME="C2"
gnome-terminal -t "C2" -x bash -c "./qbrun.exe transaction -from 1TJhYPkgZ7xngyu137D4VHvn2EySuwbEh -to 1CG9GcxF2BT1rjxwoMHSLtRP9RTCsSkRyH -amount 2;exec bash"

export NODE_NAME="C3"
gnome-terminal -t "C3" -x bash -c "./qbrun.exe transaction -from 1B8Qa4wTDLw4ywRHQ87Pp4GkypRrDyWbKm -to 1TJhYPkgZ7xngyu137D4VHvn2EySuwbEh -amount 2;exec bash"

export NODE_NAME="C4"
gnome-terminal -t "C4" -x bash -c "./qbrun.exe transaction -from 1sJT4CzXPViuT59FY6iQ8XpEhTYNz1BpK -to 1TJhYPkgZ7xngyu137D4VHvn2EySuwbEh -amount 2;exec bash"

