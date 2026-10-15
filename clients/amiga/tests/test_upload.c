#include "amirender_upload.h"

#include <stdio.h>
#include <string.h>

int main(void)
{
    char buffer[256];
    const unsigned char scene[] = "camera {}\n";
    const char *expected =
        "{\"type\":\"UPLOAD\",\"name\":\"scene.pov\",\"data\":\"Y2FtZXJhIHt9Cg==\"}\n";

    if (amirender_build_upload(buffer, sizeof(buffer), "scene.pov", scene, sizeof(scene) - 1) < 0) return 1;
    if (strcmp(buffer, expected) != 0) {
        fprintf(stderr, "unexpected upload: %s", buffer);
        return 2;
    }
    if (amirender_build_upload(buffer, sizeof(buffer), "../scene.pov", scene, sizeof(scene) - 1) >= 0) return 3;
    if (amirender_build_upload(buffer, 16, "scene.pov", scene, sizeof(scene) - 1) >= 0) return 4;
    return 0;
}
