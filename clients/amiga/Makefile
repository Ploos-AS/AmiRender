CC ?= m68k-amigaos-gcc
CFLAGS ?= -O2 -Wall -Wextra
CPPFLAGS ?= -Iinclude
TARGET ?= amirender

SOURCES = \
	src/submit.c \
	src/client.c \
	src/bsdsocket_amiga.c \
	src/main_amiga.c

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(SOURCES)
	$(CC) $(CPPFLAGS) $(CFLAGS) $(SOURCES) -o $@

clean:
	rm -f $(TARGET)
