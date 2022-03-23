#!/bin/bash

echo "run node server!"
set NODE_NAME

export NODE_NAME="P1"
gnome-terminal -t "BC1" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P2"
gnome-terminal -t "BC2" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P3"
gnome-terminal -t "BC3" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P4"
gnome-terminal -t "BC4" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P5"
gnome-terminal -t "BC5" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P6"
gnome-terminal -t "BC6" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P7"
gnome-terminal -t "BC7" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P8"
gnome-terminal -t "BC8" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P9"
gnome-terminal -t "BC9" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P10"
gnome-terminal -t "BC10" -x bash -c "./qbrun.exe startnode;exec bash"

:<<!
export NODE_NAME="P11"
gnome-terminal -t "BC11" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P12"
gnome-terminal -t "BC12" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P13"
gnome-terminal -t "BC13" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P14"
gnome-terminal -t "BC14" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P15"
gnome-terminal -t "BC15" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P16"
gnome-terminal -t "BC16" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P17"
gnome-terminal -t "BC17" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P18"
gnome-terminal -t "BC18" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P19"
gnome-terminal -t "BC19" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P20"
gnome-terminal -t "BC20" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P21"
gnome-terminal -t "BC21" -x bash -c "./qbrun.exe startnode;exec bash"

export NODE_NAME="P22"
gnome-terminal -t "BC22" -x bash -c "./qbrun.exe startnode;exec bash"
!
:<<!
-t 为打开终端的标题，便于区分。
最后的exec bash;是让打开的终端在执行完脚本后不关闭。
!


 
