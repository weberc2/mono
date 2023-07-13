#ifndef COPY_H
#define COPY_H

#include "reader.h"
#include "writer.h"

io_result copy_buf(writer dst, reader src, str buf);
io_result copy(writer dst, reader src);

#endif // COPY_H