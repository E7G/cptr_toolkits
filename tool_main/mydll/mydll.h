#ifndef MYDLL_H
#define MYDLL_H

#ifdef MYDLL_EXPORTS
#define MYDLL_API __declspec(dllexport)
#else
#define MYDLL_API __declspec(dllimport)
#endif

#ifdef __cplusplus
extern "C" {
#endif

MYDLL_API int unzipFile(const char* archive, const char* outputFolder);
MYDLL_API int compressFolder(const char* folder, const char* archive);
MYDLL_API int runCmd(const char* cmd);

#ifdef __cplusplus
}
#endif

#endif // MYDLL_H