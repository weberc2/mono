#include "test.h"
#include "io/copy.h"
#include "io/str_reader.h"
#include "string/string.h"
#include <stdbool.h>

bool test_str_reader()
{
    test_init("test_str_reader");

    char data[] = "helloworld";
    str source;
    str_init(&source, data, sizeof(data) - 1);

    char buf[6] = ".....";
    str buffer;
    str_init(&buffer, buf, sizeof(buf) - 1);

    str_reader br;
    str_reader_init(&br, source);

    reader r;
    str_reader_to_reader(&br, &r);

    errors errs;
    errors_init(&errs);
    TEST_DEFER(errors_drop, &errs);

    size_t nr = reader_read(r, buffer, &errs);

    if (nr != buffer.len)
    {
        return test_fail("nr: wanted `%zu`; found `%zu`", buffer.len, nr);
    }

    str wanted;
    str_init(&wanted, "hello", 5);
    if (!str_eq(wanted, buffer))
    {
        return test_fail(
            "data: wanted `%s`; found `%s`",
            wanted.data,
            buffer.data);
    }
    if (errors_len(&errs) > 0)
    {
        return test_fail("found `%d` unexpected errors", errors_len(&errs));
    }

    // Read a second time to get the rest of the data
    nr = reader_read(r, buffer, &errs);

    if (nr != buffer.len)
    {
        return test_fail("nr: wanted `%zu`; found `%zu`", buffer.len, nr);
    }

    str_init(&wanted, "world", 5);
    if (!str_eq(wanted, buffer))
    {
        return test_fail(
            "data: wanted `%s`; found `%s`",
            wanted.data,
            buffer.data);
    }
    if (errors_len(&errs) > 0)
    {
        return test_fail("found `%d` unexpected errors", errors_len(&errs));
    }

    // read a third time to get the eof
    nr = reader_read(r, buffer, &errs);
    if (nr != 0)
    {
        return test_fail("nr: wanted `0`; found `%zu`", nr);
    }
    if (errors_len(&errs) > 0)
    {
        return test_fail("found `%d` unexpected errors", errors_len(&errs));
    }

    return test_success();
}

bool test_copy()
{
    test_init("test_copy");

    vector deferables;
    vector_init(&deferables, sizeof(struct deferable));
    deferables_push(&deferables, &deferables, (defer_func)vector_drop);
    TEST_DEFER(defer_many, &deferables);

    char srcdata[] = "helloworld";
    str src;
    str_init(&src, srcdata, sizeof(srcdata) - 1);

    string dst;
    string_init(&dst);
    deferables_push(&deferables, (void *)&dst, (defer_func)string_drop);

    errors errs;
    errors_init(&errs);
    deferables_push(&deferables, &errs, (defer_func)errors_drop);

    str_reader str_reader;
    str_reader_init(&str_reader, src);

    reader r;
    str_reader_to_reader(&str_reader, &r);

    writer w;
    writer_from_string(&w, &dst);

    copy(w, r, &errs);

    str dstslice;
    string_borrow(&dst, &dstslice);
    if (!str_eq(src, dstslice))
    {
        return test_fail("wanted `%s`; found `%s`", src.data, dst.data);
    }

    return test_success();
}

bool io_tests()
{
    return test_str_reader() && test_copy();
}