#include "sqlite3.h"

#include <assert.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>

void *goFsOpen(void *pfs, const char *name);
int goFileClose(void *fileAppData);
int goFileRead(void *fileAppData, char *buf, int n, int64_t offset);
int goFileWrite(void *fileAppData, const char *buf, int n, int64_t offset);
int64_t goFileSize(void *fileAppData);

/*
** Maximum pathname length supported by the fs backend.
*/
#define BLOCKSIZE 512

typedef struct fs_file fs_file;
struct fs_file
{
    sqlite3_file base;
    void *appData;
};

/*
** Method declarations for fs_file.
*/
static int fsClose(sqlite3_file *);
static int fsRead(sqlite3_file *, void *, int iAmt, sqlite3_int64 iOfst);
static int fsWrite(sqlite3_file *, const void *, int iAmt, sqlite3_int64 iOfst);
static int fsTruncate(sqlite3_file *, sqlite3_int64 size);
static int fsSync(sqlite3_file *, int flags);
static int fsFileSize(sqlite3_file *, sqlite3_int64 *pSize);
static int fsLock(sqlite3_file *, int);
static int fsUnlock(sqlite3_file *, int);
static int fsCheckReservedLock(sqlite3_file *, int *pResOut);
static int fsFileControl(sqlite3_file *, int op, void *pArg);
static int fsSectorSize(sqlite3_file *);
static int fsDeviceCharacteristics(sqlite3_file *);

/*
** Method declarations for fs_vfs.
*/
static int fsOpen(sqlite3_vfs *, const char *, sqlite3_file *, int, int *);
static int fsDelete(sqlite3_vfs *, const char *zName, int syncDir);
static int fsAccess(sqlite3_vfs *, const char *zName, int flags, int *);
static int fsFullPathname(sqlite3_vfs *, const char *zName, int nOut, char *zOut);
static void *fsDlOpen(sqlite3_vfs *, const char *zFilename);
static void fsDlError(sqlite3_vfs *, int nByte, char *zErrMsg);
static void (*fsDlSym(sqlite3_vfs *, void *, const char *zSymbol))(void);
static void fsDlClose(sqlite3_vfs *, void *);
static int fsRandomness(sqlite3_vfs *, int nByte, char *zOut);
static int fsSleep(sqlite3_vfs *, int microseconds);
static int fsCurrentTime(sqlite3_vfs *, double *);

static sqlite3_vfs fs_vfs = {
    1,              /* iVersion */
    0,              /* szOsFile */
    1024,           /* mxPathname */
    0,              /* pNext */
    0,              /* zName */
    0,              /* pAppData */
    fsOpen,         /* xOpen */
    fsDelete,       /* xDelete */
    fsAccess,       /* xAccess */
    fsFullPathname, /* xFullPathname */
    fsDlOpen,       /* xDlOpen */
    fsDlError,      /* xDlError */
    fsDlSym,        /* xDlSym */
    fsDlClose,      /* xDlClose */
    fsRandomness,   /* xRandomness */
    fsSleep,        /* xSleep */
    fsCurrentTime,  /* xCurrentTime */
    0               /* xCurrentTimeInt64 */
};

static sqlite3_io_methods fs_io_methods = {
    1,                       /* iVersion */
    fsClose,                 /* xClose */
    fsRead,                  /* xRead */
    fsWrite,                 /* xWrite */
    fsTruncate,              /* xTruncate */
    fsSync,                  /* xSync */
    fsFileSize,              /* xFileSize */
    fsLock,                  /* xLock */
    fsUnlock,                /* xUnlock */
    fsCheckReservedLock,     /* xCheckReservedLock */
    fsFileControl,           /* xFileControl */
    fsSectorSize,            /* xSectorSize */
    fsDeviceCharacteristics, /* xDeviceCharacteristics */
    0,                       /* xShmMap */
    0,                       /* xShmLock */
    0,                       /* xShmBarrier */
    0                        /* xShmUnmap */
};

/* Useful macros used in several places */
#define MIN(x, y) ((x) < (y) ? (x) : (y))
#define MAX(x, y) ((x) > (y) ? (x) : (y))

/*
** Close an fs-file.
*/
static int fsClose(sqlite3_file *pFile)
{
    fs_file *p = (fs_file *)pFile;
    return goFileClose(p->appData);
}

/*
** Read data from an fs-file.
*/
static int fsRead(
    sqlite3_file *pFile,
    void *zBuf,
    int iAmt,
    sqlite_int64 iOfst)
{
    fs_file *p = (fs_file *)pFile;
    return goFileRead(p->appData, zBuf, iAmt, iOfst);
}

/*
** Write data to an fs-file.
*/
static int fsWrite(
    sqlite3_file *pFile,
    const void *zBuf,
    int iAmt,
    sqlite_int64 iOfst)
{
    fs_file *p = (fs_file *)pFile;
    int rc = goFileWrite(p->appData, zBuf, iAmt, iOfst);
    if (rc != 0)
    {
        return SQLITE_IOERR_WRITE;
    }
    return rc;
}

/*
** Truncate an fs-file.
*/
static int fsTruncate(sqlite3_file *pFile, sqlite_int64 size)
{
    return SQLITE_OK;
}

/*
** Sync an fs-file.
*/
static int fsSync(sqlite3_file *pFile, int flags)
{
    return SQLITE_OK;
}

/*
** Return the current file-size of an fs-file.
*/
static int fsFileSize(sqlite3_file *pFile, sqlite_int64 *pSize)
{
    fs_file *p = (fs_file *)pFile;
    *pSize = goFileSize(p->appData);
    return SQLITE_OK;
}

/*
** Lock an fs-file.
*/
static int fsLock(sqlite3_file *pFile, int eLock)
{
    return SQLITE_OK;
}

/*
** Unlock an fs-file.
*/
static int fsUnlock(sqlite3_file *pFile, int eLock)
{
    return SQLITE_OK;
}

/*
** Check if another file-handle holds a RESERVED lock on an fs-file.
*/
static int fsCheckReservedLock(sqlite3_file *pFile, int *pResOut)
{
    *pResOut = 0;
    return SQLITE_OK;
}

/*
** File control method. For custom operations on an fs-file.
*/
static int fsFileControl(sqlite3_file *pFile, int op, void *pArg)
{
    if (op == SQLITE_FCNTL_PRAGMA)
        return SQLITE_NOTFOUND;
    return SQLITE_OK;
}

/*
** Return the sector-size in bytes for an fs-file.
*/
static int fsSectorSize(sqlite3_file *pFile)
{
    return BLOCKSIZE;
}

/*
** Return the device characteristic flags supported by an fs-file.
*/
static int fsDeviceCharacteristics(sqlite3_file *pFile)
{
    return SQLITE_IOCAP_BATCH_ATOMIC | SQLITE_IOCAP_POWERSAFE_OVERWRITE;
}

/*
** Open an fs file handle.
*/
static int fsOpen(
    sqlite3_vfs *pVfs,
    const char *zName,
    sqlite3_file *pFile,
    int flags,
    int *pOutFlags)
{
    printf("open %s\n", zName);
    void *fileAppData = NULL;
    pFile->pMethods = &fs_io_methods;
    fs_file *p = (fs_file *)pFile;
    if ((flags & SQLITE_OPEN_MAIN_DB) == 0)
    {
        return SQLITE_IOERR;
    }
    fileAppData = goFsOpen(pVfs->pAppData, zName);
    if (fileAppData == NULL)
    {
        return SQLITE_ERROR;
    }
    p->appData = fileAppData;
    return SQLITE_OK;
}

/*
** Delete the file located at zPath. If the dirSync argument is true,
** ensure the file-system modifications are synced to disk before
** returning.
*/
static int fsDelete(sqlite3_vfs *pVfs, const char *zPath, int dirSync)
{
    printf("delete %s\n", zPath);
    return SQLITE_OK;
}

/*
** Test for access permissions. Return true if the requested permission
** is available, or false otherwise.
*/
static int fsAccess(
    sqlite3_vfs *pVfs,
    const char *zPath,
    int flags,
    int *pResOut)
{
    *pResOut = 0;
    return SQLITE_OK;
}

/*
** Populate buffer zOut with the full canonical pathname corresponding
** to the pathname in zPath. zOut is guaranteed to point to a buffer
** of at least (FS_MAX_PATHNAME+1) bytes.
*/
static int fsFullPathname(
    sqlite3_vfs *pVfs, /* Pointer to vfs object */
    const char *zPath, /* Possibly relative input path */
    int nOut,          /* Size of output buffer in bytes */
    char *zOut         /* Output buffer */
)
{
    sqlite3_snprintf(nOut, zOut, "%s", zPath);
    return SQLITE_OK;
}

/*
** Open the dynamic library located at zPath and return a handle.
*/
static void *fsDlOpen(sqlite3_vfs *pVfs, const char *zPath)
{
    return SQLITE_OK;
}

/*
** Populate the buffer zErrMsg (size nByte bytes) with a human readable
** utf-8 string describing the most recent error encountered associated 
** with dynamic libraries.
*/
static void fsDlError(sqlite3_vfs *pVfs, int nByte, char *zErrMsg)
{
    return;
}

/*
** Return a pointer to the symbol zSymbol in the dynamic library pHandle.
*/
static void (*fsDlSym(sqlite3_vfs *pVfs, void *pH, const char *zSym))(void)
{
    return NULL;
}

/*
** Close the dynamic library handle pHandle.
*/
static void fsDlClose(sqlite3_vfs *pVfs, void *pHandle)
{
    return;
}

/*
** Populate the buffer pointed to by zBufOut with nByte bytes of 
** random data.
*/
static int fsRandomness(sqlite3_vfs *pVfs, int nByte, char *zBufOut)
{
    memset(zBufOut, 0, nByte);
    return SQLITE_OK;
}

/*
** Sleep for nMicro microseconds. Return the number of microseconds 
** actually slept.
*/
static int fsSleep(sqlite3_vfs *pVfs, int nMicro)
{
    return SQLITE_OK;
}

/*
** Return the current time as a Julian Day number in *pTimeOut.
*/
static int fsCurrentTime(sqlite3_vfs *pVfs, double *pTimeOut)
{
    *pTimeOut = 0;
    return SQLITE_OK;
}

/*
** This procedure registers the fs vfs with SQLite. If the argument is
** true, the fs vfs becomes the new default vfs. It is the only publicly
** available function in this file.
*/
int fs_register(char *name, void *appData)
{
    sqlite3_vfs *vfs = malloc(sizeof(sqlite3_vfs));
    *vfs = fs_vfs;
    vfs->szOsFile = sizeof(fs_file);
    vfs->pAppData = appData;
    vfs->zName = strdup(name);
    return sqlite3_vfs_register(vfs, 0);
}