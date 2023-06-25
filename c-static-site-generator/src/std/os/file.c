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

size_t file_write(file f, str buf, result *res)
{
    if (feof(f.handle) != 0)
    {
        return 0;
    }
    size_t nw = fwrite(buf.data, 1, buf.len, f.handle);
    if (ferror(f.handle))
    {
        *res = result_err(errno_error(errno));
    }
    else
    {
        *res = result_ok();
    }
    return nw;
}

size_t file_read(file f, str buf, result *res)
{
    if (feof(f.handle) != 0)
    {
        return 0;
    }
    size_t nr = fread(buf.data, 1, buf.len, f.handle);
    if (ferror(f.handle))
    {
        *res = result_err(errno_error(errno));
    }
    else
    {
        *res = result_ok();
    }
    return nr;
}

result file_close(file f)
{
    if (fclose(f.handle) == 0)
    {
        return result_ok();
    }

    return result_err(errno_error(errno));
}

reader file_reader(file f)
{
    return reader_new((void *)f.handle, (read_func)file_read);
}

writer file_writer(file f)
{
    return writer_new((void *)f.handle, (write_func)file_write);
}