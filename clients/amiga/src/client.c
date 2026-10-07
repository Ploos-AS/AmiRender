#include "amirender_submit.h"
#include "amirender_transport.h"
#include "amirender_upload.h"

#include <stdio.h>
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

static int base64_value(char ch)
{
    if (ch >= 'A' && ch <= 'Z') return ch - 'A';
    if (ch >= 'a' && ch <= 'z') return ch - 'a' + 26;
    if (ch >= '0' && ch <= '9') return ch - '0' + 52;
    if (ch == '+') return 62;
    if (ch == '/') return 63;
    return -1;
}

int amirender_download_asset(
    struct amirender_transport *transport,
    const char *asset,
    unsigned char *data,
    size_t data_size,
    size_t *received_size)
{
    char request[1200];
    char reply[8192];
    const char *key = "\"data\":\"";
    const char *p;
    const char *end;
    size_t out = 0;
    int received;

    if (transport == NULL || transport->send == NULL || transport->receive == NULL ||
        asset == NULL || data == NULL || received_size == NULL) return -1;
    if (strchr(asset, '\"') != NULL || strchr(asset, '\\') != NULL) return -1;
    if (snprintf(request, sizeof(request), "{\"type\":\"DOWNLOAD\",\"asset\":\"%s\"}\n", asset) < 0 ||
        strlen(request) >= sizeof(request)) return -1;
    if (transport->send(transport->context, request, strlen(request)) != (int)strlen(request)) return -1;
    received = transport->receive(transport->context, reply, sizeof(reply) - 1);
    if (received <= 0 || (size_t)received >= sizeof(reply)) return -1;
    reply[received] = '\0';
    if (strstr(reply, "\"type\":\"DATA\"") == NULL) return -1;
    p = strstr(reply, key);
    if (p == NULL) return -1;
    p += strlen(key);
    end = strchr(p, '\"');
    if (end == NULL) return -1;

    while (p < end) {
        int a, b, d, e;
        unsigned int v;
        if (end - p < 4) return -1;
        a = base64_value(p[0]); b = base64_value(p[1]);
        d = p[2] == '=' ? 0 : base64_value(p[2]);
        e = p[3] == '=' ? 0 : base64_value(p[3]);
        if (a < 0 || b < 0 || d < 0 || e < 0) return -1;
        v = ((unsigned int)a << 18) | ((unsigned int)b << 12) |
            ((unsigned int)d << 6) | (unsigned int)e;
        if (out >= data_size) return -1;
        data[out++] = (unsigned char)(v >> 16);
        if (p[2] != '=') {
            if (out >= data_size) return -1;
            data[out++] = (unsigned char)(v >> 8);
        }
        if (p[3] != '=') {
            if (out >= data_size) return -1;
            data[out++] = (unsigned char)v;
        }
        p += 4;
    }
    *received_size = out;
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
