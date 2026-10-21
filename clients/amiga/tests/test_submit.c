#include "amirender_submit.h"

#include <stdio.h>
#include <string.h>

int main(void)
{
    char buffer[512];
    struct amirender_job job = {
        "amiga-0001", "povray", "DH0:Scenes/demo.pov",
        "RAM:frame.png", 0, 320, 256
    };
    const char *expected =
        "{\"type\":\"SUBMIT\",\"job\":{\"id\":\"amiga-0001\",\"engine\":\"povray\","
        "\"scene\":\"DH0:Scenes/demo.pov\",\"output\":\"RAM:frame.png\","
        "\"frame\":0,\"width\":320,\"height\":256}}\n";

    if (amirender_build_submit(buffer, sizeof(buffer), &job) < 0) {
        return 1;
    }
    if (strcmp(buffer, expected) != 0) {
        fprintf(stderr, "unexpected message: %s", buffer);
        return 2;
    }
    job.scene = "bad\"scene.pov";
    if (amirender_build_submit(buffer, sizeof(buffer), &job) >= 0) {
        return 3;
    }
    return 0;
}
