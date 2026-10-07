#include "amirender_submit.h"
#include "amirender_transport.h"
#include "amirender_upload.h"

#include <string.h>

int amirender_upload_asset(
    struct amirender_transport *transport,
    const char *name,
    const unsigned char *data,
    size_t data_size,
    char *asset,
    size_t asset_size)
{
    char request[8192];
    char reply[1024];
    const char *key = "\"asset\":\"";
    const char *start;
    const char *end;
    size_t length;
    int request_length;
    int received;

    if (transport == NULL || transport->send == NULL || transport->receive == NULL ||
        asset == NULL || asset_size < 2) {
        return -1;
    }
    request_length = amirender_build_upload(request, sizeof(request), name, data, data_size);
    if (request_length < 0) {
        return -1;
    }
    if (transport->send(transport->context, request, (size_t)request_length) != request_length) {
        return -1;
    }
    received = transport->receive(transport->context, reply, sizeof(reply) - 1);
    if (received <= 0 || (size_t)received >= sizeof(reply)) {
        return -1;
    }
    reply[received] = '\0';
    if (strstr(reply, "\"type\":\"STAGED\"") == NULL) {
        return -1;
    }
    start = strstr(reply, key);
    if (start == NULL) {
        return -1;
    }
    start += strlen(key);
    end = strchr(start, '"');
    if (end == NULL) {
        return -1;
    }
    length = (size_t)(end - start);
    if (length == 0 || length >= asset_size) {
        return -1;
    }
    memcpy(asset, start, length);
    asset[length] = '\0';
    return 0;
}

int amirender_submit_job(
    struct amirender_transport *transport,
    const struct amirender_job *job,
    char *reply,
    size_t reply_size)
{
    char request[1024];
    int length;
    int received;

    if (transport == NULL || transport->send == NULL || transport->receive == NULL ||
        reply == NULL || reply_size < 2) {
        return -1;
    }

    length = amirender_build_submit(request, sizeof(request), job);
    if (length < 0) {
        return -1;
    }
    if (transport->send(transport->context, request, (size_t)length) != length) {
        return -1;
    }

    received = transport->receive(transport->context, reply, reply_size - 1);
    if (received <= 0 || (size_t)received >= reply_size) {
        return -1;
    }
    reply[received] = '\0';

    if (strstr(reply, "\"type\":\"COMPLETE\"") == NULL) {
        return -1;
    }
    return 0;
}
