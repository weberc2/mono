#include "std/os/file.h"
#include "std/os/errno.h"
#include "std/os/errno.h"
#include <errno.h>

static void __attribute((constructor)) init()
{
    file *f = (file *)(&file_stdout);
    f->handle = stdout;

    f = (file *)(&file_stderr);
    f->handle = stderr;

    f = (file *)(&file_stdin);
    f->handle = stdin;
}

io_result file_write(file f, str buf)
{
    if (feof(f.handle) != 0)
    {
        return IO_RESULT_OK(0);
    }

    size_t nw = fwrite(buf.data, 1, buf.len, f.handle);
    return IO_RESULT(nw, ferror(f.handle) ? errno_error(errno) : ERROR_NULL);
}

io_result file_read(file f, str buf)
{
    if (feof(f.handle) != 0)
    {
        return IO_RESULT_OK(0);
    }
    size_t nr = fread(buf.data, 1, buf.len, f.handle);
    return IO_RESULT(nr, ferror(f.handle) ? errno_error(errno) : ERROR_NULL);
}

error file_close(file f)
{
    if (fclose(f.handle) == 0)
    {
        return ERROR_NULL;
    }

    return errno_error(errno);
}

reader file_reader(file f)
{
    return reader_new((void *)f.handle, (read_func)file_read);
}

writer file_writer(file f)
{
    return writer_new((void *)f.handle, (write_func)file_write);
}