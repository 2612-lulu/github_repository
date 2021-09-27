#!/bin/bash

echo "run node server!"
set NODE_NAME
export NODE_NAME="P1"
#echo $NODE_NAME
echo "run node P1!"
gnome-terminal -t "P1" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P2"
echo "run node P2!"
gnome-terminal -t "P2" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P3"
echo "run node P3!"
gnome-terminal -t "P3" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P4"
echo "run node P4!"
gnome-terminal -t "P4" -x bash -c "./qbrun.exe startnode;exec bash"
:<<!
-t 为打开终端的标题，便于区分。
最后的exec bash;是让打开的终端在执行完脚本后不关闭。
!


 
