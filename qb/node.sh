#!/bin/bash

echo "run node server!"
set NODE_NAME
export NODE_NAME="P1"
gnome-terminal -t "P1" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P2"
gnome-terminal -t "P2" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P3"
gnome-terminal -t "P3" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P4"
gnome-terminal -t "P4" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P5"
gnome-terminal -t "P5" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P6"
gnome-terminal -t "P6" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P7"
gnome-terminal -t "P7" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P8"
gnome-terminal -t "P8" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P9"
gnome-terminal -t "P9" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P10"
gnome-terminal -t "P10" -x bash -c "./qbrun.exe startnode;exec bash"

:<<!
export NODE_NAME="P11"
gnome-terminal -t "P11" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P12"
gnome-terminal -t "P12" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P13"
gnome-terminal -t "P13" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P14"
gnome-terminal -t "P14" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P15"
gnome-terminal -t "P15" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P16"
gnome-terminal -t "P16" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P17"
gnome-terminal -t "P17" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P18"
gnome-terminal -t "P18" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P19"
gnome-terminal -t "P19" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P20"
gnome-terminal -t "P20" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P21"
gnome-terminal -t "P21" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P22"
gnome-terminal -t "P22" -x bash -c "./qbrun.exe startnode;exec bash"
!

:<<!
-t 为打开终端的标题，便于区分。
最后的exec bash;是让打开的终端在执行完脚本后不关闭。
!


 
