CFLAGS=-Wall --std=c99
TARGETS=success failure infinite

all: $(TARGETS)

success: success.c
	gcc $(CFLAGS) -o $@ $< -lpthread

failure: failure.c
	gcc $(CFLAGS) -o $@ $< -lpthread

infinite: infinite.c
	gcc $(CFLAGS) -o $@ $< -lpthread

clean:
	rm -f $(TARGETS)
