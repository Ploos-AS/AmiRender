#include "amirender_upload.h"

#include <stdio.h>
#include <string.h>

static const char base64_table[] =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

static int safe_name(const char *name)
{
    const unsigned char *p = (const unsigned char *)name;

    if (name == NULL || *name == '\0' || strcmp(name, ".") == 0 || strcmp(name, "..") == 0) {
        return 0;
    }
    while (*p != '\0') {
        if (*p < 0x20 || *p == '"' || *p == '\\' || *p == '/' || *p == ':') {
            return 0;
        }
        ++p;
    }
    return 1;
}

int amirender_build_upload(
    char *buffer,
    size_t size,
    const char *name,
    const unsigned char *data,
    size_t data_size)
{
    size_t encoded_size;
    size_t required;
    size_t i;
    size_t out;
    int prefix;

    if (buffer == NULL || size == 0 || !safe_name(name) || data == NULL || data_size == 0) {
        return -1;
    }
    if (data_size > (size_t)-1 - 2) {
        return -1;
    }
    encoded_size = ((data_size + 2) / 3) * 4;
    prefix = snprintf(buffer, size, "{\"type\":\"UPLOAD\",\"name\":\"%s\",\"data\":\"", name);
    if (prefix < 0 || (size_t)prefix >= size) {
        return -1;
    }
    required = (size_t)prefix + encoded_size + 3;
    if (required > size) {
        return -1;
    }

    out = (size_t)prefix;
    for (i = 0; i < data_size; i += 3) {
        unsigned long v = (unsigned long)data[i] << 16;
        size_t remain = data_size - i;
        if (remain > 1) v |= (unsigned long)data[i + 1] << 8;
        if (remain > 2) v |= data[i + 2];

        buffer[out++] = base64_table[(v >> 18) & 63];
        buffer[out++] = base64_table[(v >> 12) & 63];
        buffer[out++] = remain > 1 ? base64_table[(v >> 6) & 63] : '=';
        buffer[out++] = remain > 2 ? base64_table[v & 63] : '=';
    }
    buffer[out++] = '"';
    buffer[out++] = '}';
    buffer[out++] = '\n';
    buffer[out] = '\0';
    return (int)out;
}
