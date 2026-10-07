#ifndef AMIRENDER_TRANSPORT_H
#define AMIRENDER_TRANSPORT_H

#include <stddef.h>

struct amirender_job;

struct amirender_transport {
    void *context;
    int (*send)(void *context, const char *data, size_t length);
    int (*receive)(void *context, char *buffer, size_t size);
};

int amirender_upload_asset(
    struct amirender_transport *transport,
    const char *name,
    const unsigned char *data,
    size_t data_size,
    char *asset,
    size_t asset_size);

int amirender_download_asset(
    struct amirender_transport *transport,
    const char *asset,
    unsigned char *data,
    size_t data_size,
    size_t *received_size);

int amirender_upload_begin(
    struct amirender_transport *transport,
    const char *name,
    size_t total_size,
    char *asset,
    size_t asset_size);

int amirender_upload_chunk(
    struct amirender_transport *transport,
    const char *asset,
    size_t offset,
    const unsigned char *data,
    size_t data_size);

int amirender_download_chunk(
    struct amirender_transport *transport,
    const char *asset,
    size_t offset,
    unsigned char *data,
    size_t data_size,
    size_t *received_size,
    int *eof);

int amirender_submit_job(
    struct amirender_transport *transport,
    const struct amirender_job *job,
    char *reply,
    size_t reply_size);

#endif
