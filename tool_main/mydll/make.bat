set PATH=%PATH%;E:\Dev-Cpp\tdm-gcc\bin
gcc -shared -o mydll.dll mydll.c -Wl,--out-implib,libmydll.a -Wl,--subsystem,windows
pause