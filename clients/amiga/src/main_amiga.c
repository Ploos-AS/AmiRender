#include "amirender_bsdsocket.h"
#include "amirender_submit.h"
#include "amirender_transport.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define AMIRENDER_PORT 6800
#define REPLY_SIZE 1024
#define ASSET_SIZE 1024
#define CHUNK_SIZE 2048

static const char *base_name(const char *path)
{
    const char *name = path;
    const char *p;

    for (p = path; *p != '\0'; ++p) {
        if (*p == '/' || *p == ':') {
            name = p + 1;
        }
    }
    return name;
}

static int scene_size(const char *path, size_t *length)
{
    FILE *file;
    long size;

    file = fopen(path, "rb");
    if (file == NULL) return -1;
    if (fseek(file, 0, SEEK_END) != 0) {
        fclose(file);
        return -1;
    }
    size = ftell(file);
    fclose(file);
    if (size <= 0) return -1;
    *length = (size_t)size;
    return 0;
}

static int upload_file(
    struct amirender_transport *transport, const char *path,
    char *asset, size_t asset_size)
{
    FILE *file;
    unsigned char buffer[CHUNK_SIZE];
    size_t total, offset = 0, count;

    if (scene_size(path, &total) != 0) return -1;
    if (amirender_upload_begin(
            transport, base_name(path), total, asset, asset_size) != 0) return -1;
    file = fopen(path, "rb");
    if (file == NULL) return -1;
    while ((count = fread(buffer, 1, sizeof(buffer), file)) > 0) {
        if (amirender_upload_chunk(transport, asset, offset, buffer, count) != 0) {
            fclose(file);
            return -1;
        }
        offset += count;
    }
    if (ferror(file) || offset != total) {
        fclose(file);
        return -1;
    }
    if (fclose(file) != 0) return -1;
    return amirender_upload_end(transport, asset, total);
}

static int extract_output(const char *reply, char *output, size_t output_size)
{
    const char *key = "\"output\":\"";
    const char *start = strstr(reply, key);
    const char *end;
    size_t length;

    if (start == NULL) return -1;
    start += strlen(key);
    end = strchr(start, '\"');
    if (end == NULL) return -1;
    length = (size_t)(end - start);
    if (length == 0 || length >= output_size) return -1;
    memcpy(output, start, length);
    output[length] = '\0';
    return 0;
}

static int download_file(
    struct amirender_transport *transport, const char *asset, const char *path,
    size_t *total_size)
{
    FILE *file;
    unsigned char buffer[CHUNK_SIZE];
    size_t offset = 0, count;
    int eof = 0;

    file = fopen(path, "wb");
    if (file == NULL) return -1;
    while (!eof) {
        if (amirender_download_chunk(
                transport, asset, offset, buffer, sizeof(buffer), &count, &eof) != 0) {
            fclose(file);
            remove(path);
            return -1;
        }
        if (count > 0 && fwrite(buffer, 1, count, file) != count) {
            fclose(file);
            remove(path);
            return -1;
        }
        offset += count;
        if (count == 0 && !eof) {
            fclose(file);
            remove(path);
            return -1;
        }
    }
    if (fclose(file) != 0) {
        remove(path);
        return -1;
    }
    *total_size = offset;
    return 0;
}

static void usage(const char *program)
{
    fprintf(stderr,
        "Usage: %s HOST SCENE OUTPUT [WIDTH HEIGHT]\n"
        "Example: %s 192.168.1.50 DH0:Scenes/demo.pov RAM:frame.png 320 256\n",
        program, program);
}

int main(int argc, char **argv)
{
    struct amirender_bsdsocket socket_state = {-1};
    struct amirender_transport transport = {0};
    struct amirender_job job;
    char reply[REPLY_SIZE];
    char asset[ASSET_SIZE];
    char output_asset[ASSET_SIZE];
    size_t output_size;
    int width = 0;
    int height = 0;
    int rc;

    if (argc != 4 && argc != 6) {
        usage(argv[0]);
        return 20;
    }

    if (argc == 6) {
        width = atoi(argv[4]);
        height = atoi(argv[5]);
        if (width <= 0 || height <= 0) {
            fprintf(stderr, "AmiRender: WIDTH and HEIGHT must be positive\n");
            return 20;
        }
    }

    job.id = "amiga-cli";
    job.engine = "povray";
    job.scene = argv[2];
    job.output = argv[3];
    job.frame = 0;
    job.width = width;
    job.height = height;

    if (amirender_bsdsocket_connect(
            &socket_state, argv[1], AMIRENDER_PORT, &transport) != 0) {
        fprintf(stderr, "AmiRender: cannot connect to %s:%d\n",
            argv[1], AMIRENDER_PORT);
        return 10;
    }

    rc = upload_file(&transport, argv[2], asset, sizeof(asset));
    if (rc != 0) {
        amirender_bsdsocket_close(&socket_state);
        fprintf(stderr, "AmiRender: scene upload failed\n");
        return 10;
    }

    job.scene = asset;
    rc = amirender_submit_job(&transport, &job, reply, sizeof(reply));
    if (rc != 0 || extract_output(reply, output_asset, sizeof(output_asset)) != 0) {
        amirender_bsdsocket_close(&socket_state);
        fprintf(stderr, "AmiRender: render failed\n");
        return 10;
    }

    rc = download_file(&transport, output_asset, argv[3], &output_size);
    amirender_bsdsocket_close(&socket_state);
    if (rc != 0) {
        fprintf(stderr, "AmiRender: output download failed\n");
        return 10;
    }

    printf("AmiRender: wrote %lu bytes to %s\n", (unsigned long)output_size, argv[3]);
    return 0;
}
