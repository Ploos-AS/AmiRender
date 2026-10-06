#include "amirender_bsdsocket.h"
#include "amirender_submit.h"
#include "amirender_transport.h"

#include <stdio.h>
#include <stdlib.h>

#define AMIRENDER_PORT 6800
#define REPLY_SIZE 1024

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

    rc = amirender_submit_job(&transport, &job, reply, sizeof(reply));
    amirender_bsdsocket_close(&socket_state);

    if (rc != 0) {
        fprintf(stderr, "AmiRender: render failed\n");
        return 10;
    }

    printf("%s", reply);
    return 0;
}
