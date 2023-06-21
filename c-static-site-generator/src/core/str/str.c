#include "core/str/str.h"
#include "core/math/math.h"
#include "core/panic/panic.h"
#include <string.h>

void str_init(str *s, char *data, size_t len)
{
    s->data = data;
    s->len = len;
}

void str_slice(str s, str *out, size_t start, size_t end)
{
    if (start > s.len)
    {
        panic(
            "slicing str with len `%zu`: start index `%zu` is out of "
            "bounds",
            s.len,
            start);
    }
    else if (end > s.len)
    {
        panic(
            "slicing str with len `%zu`: end index `%zu` is out of "
            "bounds",
            s.len,
            end);
    }
    else if (start > end)
    {
        panic(
            "slicing str with len `%zu`: start index `%zu` exceeds end index "
            "`%zu`",
            s.len,
            start,
            end);
    }

    out->data = s.data + start;
    out->len = end - start;
}

size_t str_copy(str dst, str src)
{
    size_t sz = min(dst.len, src.len);
    memmove(dst.data, src.data, sz);
    return sz;
}

size_t str_copy_at(str dst, str src, size_t start)
{
    str tail;
    str_slice(src, &tail, start, src.len);
    return str_copy(dst, tail);
}

bool str_eq(str lhs, str rhs)
{
    return lhs.len == rhs.len && memcmp(lhs.data, rhs.data, lhs.len) == 0;
}

size_t str_copy_to_c(char *dst, str src, size_t len)
{
    size_t copied = min(len, src.len);
    memmove(dst, src.data, copied);
    dst[copied] = '\0';
    return copied;
}

str_find_result str_find(str s, str search)
{
    if (search.len > s.len)
    {
        return (str_find_result){false, 0};
    }

    for (size_t i = 0; i < s.len; i++)
    {
        for (size_t j = 0; j < search.len; j++)
        {
            if (s.data[i + j] != search.data[j])
            {
                goto OUTER;
            }
        }

        return (str_find_result){true, i};
    OUTER:
        continue;
    }

    return (str_find_result){false, 0};
}