#ifndef COPY_H
#define COPY_H

#include "core/result/result.h"
#include "reader.h"
#include "writer.h"

size_t copy_buf(writer dst, reader src, str buf, result *res);
size_t copy(writer dst, reader src, result *res);

#endif // COPY_H