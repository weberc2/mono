#include "test.h"
#include "io/copy.h"
#include "io/str_reader.h"
#include "io/buffered_reader.h"
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

bool assert_str_eq(char *wanted_c, str found)
{
    str wanted;
    str_init(&wanted, wanted_c, strlen(wanted_c));
    if (!str_eq(wanted, found))
    {
        char *found_c = malloc(found.len + 1);
        str_copy_to_c(found_c, found, found.len);
        bool result = test_fail(
            "wanted `%s` (len: `%zu`); "
            "found `%s` (len: `%zu`)",
            wanted_c,
            wanted.len,
            found_c,
            found.len);
        free(found_c);
        return result;
    }
    return true;
}

bool assert_no_errs(errors *errs)
{
    size_t errors_count = errors_len(errs);
    if (errors_count > 0)
    {
        return test_fail("found `%zu` unexpected errors", errors_count);
    }
    return true;
}

bool assert_count(const char *ctx, size_t wanted, size_t found)
{
    if (wanted != found)
    {
        return test_fail("%s: wanted `%zu`; found `%zu`", ctx, wanted, found);
    }
    return true;
}

bool assert_read(char *wanted_c, size_t nr, str found, errors *errs)
{
    size_t len = strlen(wanted_c);
    str_slice(found, &found, 0, len);
    return assert_count("bytes read", len, nr) &&
           assert_no_errs(errs) &&
           assert_str_eq(wanted_c, found);
}

bool assert_buffered_read(
    buffered_reader *br,
    str buf,
    char *wanted_c,
    errors *errs)
{
    size_t nr = buffered_reader_read(br, buf, errs);
    return assert_read("ll", nr, buf, errs);
}

bool test_buffered_reader()
{
    test_init("test_buffered_reader");

    char srcalloc[] = "helloworld!";
    str src_str;
    str_init(&src_str, srcalloc, sizeof(srcalloc) - 1);

    str_reader src_str_reader;
    str_reader_init(&src_str_reader, src_str);

    reader src_reader;
    str_reader_to_reader(&src_str_reader, &src_reader);

    char internal_bufalloc[5] = {0};
    str internal_buffer;
    str_init(&internal_buffer, internal_bufalloc, sizeof(internal_bufalloc));

    buffered_reader br;
    buffered_reader_init(&br, src_reader, internal_buffer);

    char bufalloc[2] = {0};
    str buf;
    str_init(&buf, bufalloc, sizeof(bufalloc));

    errors errs;
    errors_init(&errs);
    TEST_DEFER(errors_drop, &errs);

#define ASSERT_BUFFERED_READ(wanted)                      \
    if (!assert_buffered_read(&br, buf, (wanted), &errs)) \
    {                                                     \
        return false;                                     \
    }
    ASSERT_BUFFERED_READ("he");
    ASSERT_BUFFERED_READ("ll");
    ASSERT_BUFFERED_READ("ow");
    ASSERT_BUFFERED_READ("or");
    ASSERT_BUFFERED_READ("ld");
    ASSERT_BUFFERED_READ("!");
#undef ASSERT_BUFFERED_READ

    return test_success();
}

bool io_tests()
{
    return test_str_reader() && test_copy() && test_buffered_reader();
}