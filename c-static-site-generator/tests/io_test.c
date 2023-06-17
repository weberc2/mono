#include "test.h"
#include "io/copy.h"
#include "io/byteslice_reader.h"
#include "bytestring/bytestring.h"
#include <stdbool.h>

bool test_byteslice_reader()
{
    test_init("test_byteslice_reader");

    char data[] = "helloworld";
    byteslice source;
    byteslice_init(&source, data, sizeof(data) - 1);

    char buf[6] = ".....";
    byteslice buffer;
    byteslice_init(&buffer, buf, sizeof(buf) - 1);

    byteslice_reader br;
    byteslice_reader_init(&br, source);

    reader r;
    byteslice_reader_to_reader(&br, &r);

    errors errs;
    errors_init(&errs);
    TEST_DEFER(errors_drop, &errs);

    size_t nr = reader_read(r, buffer, &errs);

    if (nr != buffer.len)
    {
        return test_fail("nr: wanted `%zu`; found `%zu`", buffer.len, nr);
    }

    byteslice wanted;
    byteslice_init(&wanted, "hello", 5);
    if (!byteslice_eq(wanted, buffer))
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

    byteslice_init(&wanted, "world", 5);
    if (!byteslice_eq(wanted, buffer))
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
    byteslice src;
    byteslice_init(&src, srcdata, sizeof(srcdata) - 1);

    bytestring dst;
    bytestring_init(&dst);
    deferables_push(&deferables, (void *)&dst, (defer_func)bytestring_drop);

    errors errs;
    errors_init(&errs);
    deferables_push(&deferables, &errs, (defer_func)errors_drop);

    byteslice_reader byteslice_reader;
    byteslice_reader_init(&byteslice_reader, src);

    reader r;
    byteslice_reader_to_reader(&byteslice_reader, &r);

    writer w;
    writer_from_bytestring(&w, &dst);

    copy(w, r, &errs);

    byteslice dstslice;
    bytestring_borrow(&dst, &dstslice);
    if (!byteslice_eq(src, dstslice))
    {
        return test_fail("wanted `%s`; found `%s`", src.data, dst.data);
    }

    return test_success();
}

bool io_tests()
{
    return test_byteslice_reader() && test_copy();
}