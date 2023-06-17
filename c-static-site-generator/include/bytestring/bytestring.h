#ifndef BYTESTRING_H
#define BYTESTRING_H

#include "byteslice/byteslice.h"
#include <stddef.h>

typedef struct
{
    char *data;
    size_t len;
    size_t cap;
} bytestring;

void bytestring_init(bytestring *bs);
void bytestring_drop(bytestring *bs);
void bytestring_push_raw(bytestring *bs, char *data, size_t len);
void bytestring_push_slice(bytestring *bs, byteslice src);
void bytestring_borrow(bytestring *bs, byteslice *out);
void bytestring_slice(bytestring *bs, byteslice *out, size_t start, size_t len);

#endif // BYTESTRING_H