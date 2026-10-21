#ifndef AMIRENDER_UPLOAD_H
#define AMIRENDER_UPLOAD_H

#include <stddef.h>

int amirender_build_upload(
    char *buffer,
    size_t size,
    const char *name,
    const unsigned char *data,
    size_t data_size);

#endif
