#include "amirender_bsdsocket.h"
#include "amirender_submit.h"
#include "amirender_transport.h"

#include <stdio.h>
#include <stdlib.h>

#define AMIRENDER_PORT 6800
#define REPLY_SIZE 1024
#define ASSET_SIZE 1024
#define MAX_SCENE_SIZE 6000

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

static int read_scene(const char *path, unsigned char *buffer, size_t capacity, size_t *length)
{
    FILE *file;
    long size;

    file = fopen(path, "rb");
    if (file == NULL) {
        return -1;
    }
    if (fseek(file, 0, SEEK_END) != 0) {
        fclose(file);
        return -1;
    }
    size = ftell(file);
    if (size <= 0 || (unsigned long)size > (unsigned long)capacity) {
        fclose(file);
        return -1;
    }
    if (fseek(file, 0, SEEK_SET) != 0 ||
        fread(buffer, 1, (size_t)size, file) != (size_t)size) {
        fclose(file);
        return -1;
    }
    fclose(file);
    *length = (size_t)size;
    return 0;
}

static void usage(const char *program)
{
    fprintf(stderr,
        "Usage: %s HOST SCENE OUTPUT [WIDTH HEIGHT]\n"
        "Example: %s amirender.local DH0:Scenes/demo.pov RAM:frame.png 320 256\n",
        program, program);
}

int main(int argc, char **argv)
{
    struct amirender_bsdsocket socket_state = {-1};
    struct amirender_transport transport = {0};
    struct amirender_job job;
    char reply[REPLY_SIZE];
    char asset[ASSET_SIZE];
    unsigned char scene_data[MAX_SCENE_SIZE];
    size_t scene_size;
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

    if (read_scene(argv[2], scene_data, sizeof(scene_data), &scene_size) != 0) {
        fprintf(stderr, "AmiRender: cannot read scene or scene exceeds %d bytes\n", MAX_SCENE_SIZE);
        return 10;
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

    rc = amirender_upload_asset(
        &transport, base_name(argv[2]), scene_data, scene_size, asset, sizeof(asset));
    if (rc != 0) {
        amirender_bsdsocket_close(&socket_state);
        fprintf(stderr, "AmiRender: scene upload failed\n");
        return 10;
    }

    job.scene = asset;
    rc = amirender_submit_job(&transport, &job, reply, sizeof(reply));
    amirender_bsdsocket_close(&socket_state);

    if (rc != 0) {
        fprintf(stderr, "AmiRender: render failed\n");
        return 10;
    }

    printf("%s", reply);
    return 0;
}
