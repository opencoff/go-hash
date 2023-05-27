
tool := ghash

pwd = $(shell pwd)
arch := $(shell ./build --print-arch)
bindir = ./bin/$(arch)
bin := $(bindir)/$(tool)

INSTALLDIR ?= $(HOME)/bin
installdir := $(INSTALLDIR)/$(arch)


$(bin):
	./build -s

install: $(installdir) $(bin)
	-cp -f $(bin) $(installdir)/

.PHONY: $(bin) test clean realclean $(installdir) $(INSTALLDIR)

$(installdir):
	test -d $@ || mkdir -p $@

clean realclean:
	rm -rf bin
