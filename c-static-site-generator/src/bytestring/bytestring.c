#include "bytestring/bytestring.h"
#include "math/math.h"

#include <stdlib.h> // free()
#include <string.h> // memcpy()

void bytestring_init(bytestring *bs)
{
    bs->data = NULL;
    bs->len = 0;
    bs->cap = 0;
}

void bytestring_drop(bytestring *bs)
{
    free(bs->data);
    bs->data = NULL;
    bs->len = 0;
    bs->cap = 0;
}

void bytestring_grow(bytestring *bs, size_t at_least)
{
    char *old = bs->data;
    bs->cap = max(2 * bs->cap, at_least);
    bs->data = calloc(bs->cap, 1);
    memcpy(bs->data, old, bs->len);
    free(old);
}

void bytestring_push_raw(bytestring *bs, char *data, size_t len)
{
    if (bs->cap - bs->len < len)
    {
        bytestring_grow(bs, len);
    }
    memcpy(bs->data + bs->len, data, len);
    bs->len += len;
}

void bytestring_push_slice(bytestring *bs, byteslice src)
{
    bytestring_push_raw(bs, src.data, src.len);
}

void bytestring_borrow(bytestring *bs, byteslice *out)
{
    out->data = bs->data;
    out->len = bs->len;
}

void bytestring_slice(bytestring *bs, byteslice *out, size_t start, size_t end)
{
    bytestring_borrow(bs, out);
    byteslice_slice(*out, out, start, end);
}