#include <stdio.h>

#include "core/testing/test.h"
#include "core/io/err_eof.h"
#include "core/io/str_writer.h"
#include "core/io/str_reader.h"
#include "core/io/copy.h"

typedef struct test
{
    char *name;
    reader input;
    str buf;
    str wanted;
    io_result wanted_res;
} test;

#define TEST_READER(s) READER(&STR_READER(STR(s)), str_reader_io_read)

static test tests[] = {
    {
        .name = "empty",
        .input = TEST_READER(""),
        .buf = STR_BUF(8, 0),
        .wanted = STR(""),
        .wanted_res = IO_RESULT_ERR(ERR_EOF),
    },
    {
        .name = "single-buffer",
        .input = TEST_READER("foo"),
        .buf = STR_BUF(8, 0),
        .wanted = STR("foo"),
        .wanted_res = IO_RESULT_OK(3),
    },
    {
        .name = "full-buffer",
        .input = TEST_READER("foobar"),
        .buf = STR_BUF(6, 0),
        .wanted = STR("foobar"),
        .wanted_res = IO_RESULT_OK(6),
    },
    {
        .name = "multiple-buffers",
        .input = TEST_READER("foobar"),
        .buf = STR_BUF(3, 0),
        .wanted = STR("foobar"),
        .wanted_res = IO_RESULT_OK(6),
    },
};

bool test_run(test *tc)
{
    char buf[256] = {0};
    sprintf(buf, "test_copy:%s", tc->name);
    test_init(buf);

    str_writer sw = STR_WRITER_WITH_CAP(256);
    io_result found_res = copy_buf(
        str_writer_to_writer(&sw),
        tc->input,
        tc->buf);

    if (io_result_is_ok(tc->wanted_res) && io_result_is_err(found_res))
    {
        return test_fail(
            "res: wanted `ok`; found `%s`",
            error_to_raw(found_res.err, STR_BUF(256, 0)));
    }

    if (io_result_is_err(tc->wanted_res) && io_result_is_ok(found_res))
    {
        return test_fail(
            "res: wanted `%s`; found `ok`",
            error_to_raw(tc->wanted_res.err, STR_BUF(256, 0)));
    }

    if (tc->wanted_res.size != found_res.size)
    {
        return test_fail(
            "res.size: wanted `%zu`; found `%zu`",
            tc->wanted_res.size,
            found_res.size);
    }

    str found = str_writer_data(&sw);
    if (!str_eq(tc->wanted, found))
    {
        char f[256] = {0};
        str_copy_to_c(f, found, sizeof(f));
        return test_fail(
            "wanted `%s` (len: `%zu`); found `%s` (len: `%zu`)",
            tc->wanted.data,
            tc->wanted.len,
            f,
            found.len);
    }

    return test_success();
}

bool test_copy()
{
    for (size_t i = 0; i < sizeof(tests) / sizeof(test); i++)
    {
        if (!test_run(&tests[i]))
        {
            return false;
        }
    }

    return true;
}