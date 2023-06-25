#ifndef FILE_H
#define FILE_H

#include <stdio.h>
#include "core/str/str.h"
#include "core/result/result.h"
#include "core/io/reader.h"
#include "core/io/writer.h"

typedef struct
{
    FILE *handle;
} file;

typedef enum
{
    file_mode_read,
    file_mode_write,
    file_mode_append,
    file_mode_readwrite,
    file_mode_create,
} file_mode;

size_t file_read(file f, str buf, result *res);
size_t file_write(file f, str buf, result *res);
result file_close(file f);
reader file_reader(file f);
writer file_writer(file f);

const file file_stdout;
const file file_stderr;
const file file_stdin;

#endif // FILE_H