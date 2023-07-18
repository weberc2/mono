#include "core/testing/test.h"
#include "core/io/err_eof.h"
#include "core/io/str_reader.h"

bool test_str_reader()
{
    test_init("test_str_reader");

    str source = STR("helloworld");
    str buffer = STR_ARR((char[5]){0});
    reader r = str_reader_to_reader(&STR_READER(source));
    io_result res = reader_read(r, buffer);

    if (res.size != buffer.len)
    {
        return test_fail(
            "nr: wanted `%zu`; found `%zu`",
            buffer.len,
            res.size);
    }
    ASSERT_OK(res);

    str wanted = STR("hello");
    if (!str_eq(wanted, buffer))
    {
        return test_fail(
            "data: wanted `%s`; found `%s`",
            wanted.data,
            buffer.data);
    }
    ASSERT_OK(res);

    // Read a second time to get the rest of the data
    res = reader_read(r, buffer);

    if (res.size != buffer.len)
    {
        return test_fail(
            "nr: wanted `%zu`; found `%zu`",
            buffer.len,
            res.size);
    }

    wanted = STR("world");
    if (!str_eq(wanted, buffer))
    {
        return test_fail(
            "data: wanted `%s`; found `%s`",
            wanted.data,
            buffer.data);
    }
    if (io_result_is_ok(res))
    {
        return test_fail(
            "result: unexpected `ok`; wanted `%s`",
            error_to_raw(ERR_EOF, STR_BUF(256, 0)));
    }
    if (!error_is_eof(res.err))
    {
        return test_fail(
            "result: expected `%s`; found `%s`",
            error_to_raw(ERR_EOF, STR_BUF(256, 0)),
            error_to_raw(res.err, STR_BUF(256, 0)));
    }

    return test_success();
}