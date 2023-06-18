#ifndef BYTESLICE_H
#define BYTESLICE_H

#include <stddef.h>
#include <stdbool.h>

typedef struct
{
    char *data;
    size_t len;
} byteslice;

void byteslice_init(byteslice *bs, char *data, size_t len);
void byteslice_slice(byteslice bs, byteslice *out, size_t start, size_t end);
size_t byteslice_copy(byteslice dst, byteslice src);
size_t byteslice_copy_at(byteslice dst, byteslice src, size_t start);
bool byteslice_eq(byteslice lhs, byteslice rhs);

#endif // BYTESLICE_H