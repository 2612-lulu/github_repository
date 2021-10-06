#!/bin/bash
echo "run pbft server!"
set NODE_NAME
export NODE_NAME="P1"
gnome-terminal -t "PB1" -x bash -c "./pbft.exe startPBFT;exec bash"

export NODE_NAME="P2"
gnome-terminal -t "PB2" -x bash -c "./pbft.exe startPBFT;exec bash"

export NODE_NAME="P3"
gnome-terminal -t "PB3" -x bash -c "./pbft.exe startPBFT;exec bash"

export NODE_NAME="P4"
gnome-terminal -t "PB4" -x bash -c "./pbft.exe startPBFT;exec bash"



 
