#!/bin/bash

echo "run node server!"
echo "run node P1!"
gnome-terminal -t "P1" -x bash -c "./qbrun.exe P1;exec bash"
echo "run node P2!"
gnome-terminal -t "P2" -x bash -c "./qbrun.exe P2;exec bash"
echo "run node P3!"
gnome-terminal -t "P3" -x bash -c "./qbrun.exe P3;exec bash"
echo "run node P4!"
gnome-terminal -t "P4" -x bash -c "./qbrun.exe P4;exec bash"
:<<!
-t 为打开终端的标题，便于区分。
最后的exec bash;是让打开的终端在执行完脚本后不关闭。
!


echo "run client server!"
echo "run client C1!"
gnome-terminal --window -t "C1" -x bash -c "./qbrun.exe C1 1200fromC1toC2;exec bash"
echo "run client C2!"
gnome-terminal --window -t "C2" -x bash -c "./qbrun.exe C2 1200fromC2toC1;exec bash"
echo "run client C3!"
gnome-terminal --window -t "C3" -x bash -c "./qbrun.exe C3 1200fromC3toC1;exec bash"
echo "run client C4!"
gnome-terminal --window -t "C4" -x bash -c "./qbrun.exe C4 1200fromC4toC1;exec bash"


 
 
