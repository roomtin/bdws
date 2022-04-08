CFLAGS=-Wall --std=c99
TARGETS=success failure

all: $(TARGETS)

success: success.c
	gcc $(CFLAGS) -o $@ $< -lpthread

failure: failure.c
	gcc $(CFLAGS) -o $@ $< -lpthread

clean:
	rm -f $(TARGETS)
