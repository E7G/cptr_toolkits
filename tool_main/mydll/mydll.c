#include "mydll.h"
#include <windows.h>

typedef int (__stdcall *SevenZipFunction)(int numArgs, char* args[]);

MYDLL_API int unzipFile(const char* archive, const char* outputFolder) {
    HINSTANCE hSevenZipDLL = LoadLibrary("7z.dll");
    if (hSevenZipDLL == NULL) {
        return -1;
    }

    SevenZipFunction SevenZip = (SevenZipFunction)GetProcAddress(hSevenZipDLL, "SevenZip");
    if (SevenZip == NULL) {
        FreeLibrary(hSevenZipDLL);
        return -1;
    }

    char* args[] = {
        "7z",
        "x",
        (char*)archive,
        "-ooutput_folder",
        NULL
    };

    int result = SevenZip(sizeof(args) / sizeof(args[0]) - 1, args);
    FreeLibrary(hSevenZipDLL);
    return result;
}

MYDLL_API int compressFolder(const char* folder, const char* archive) {
    HINSTANCE hSevenZipDLL = LoadLibrary("7z.dll");
    if (hSevenZipDLL == NULL) {
        return -1;
    }

    SevenZipFunction SevenZip = (SevenZipFunction)GetProcAddress(hSevenZipDLL, "SevenZip");
    if (SevenZip == NULL) {
        FreeLibrary(hSevenZipDLL);
        return -1;
    }

    char* args[] = {
        "7z",
        "a",
        (char*)archive,
        (char*)folder,
        NULL
    };

    int result = SevenZip(sizeof(args) / sizeof(args[0]) - 1, args);
    FreeLibrary(hSevenZipDLL);
    return result;
}

MYDLL_API int runCmd(const char* cmd) {
    return system(cmd);
}