include ../../Makefile.defs

all: cilium-cni

TARGET=cilium-cni

clean:
	go clean
	rm -f $(TARGET)

SOURCES := $(shell find . -name '*.go')

$(TARGET): $(SOURCES)
	go build -o $(TARGET) ./cilium-cni.go

install:
	$(INSTALL) -m 0755 -d "/usr/libexec/kubernetes/kubelet-plugins/net/exec"
	$(INSTALL) -m 0644 10-cilium-cni.conf "/usr/libexec/kubernetes/kubelet-plugins/net/exec"
	$(INSTALL) -m 0755 -d $(DESTDIR)/opt/cni/bin
	$(INSTALL) -m 0755 $(TARGET) $(DESTDIR)/opt/cni/bin
	$(INSTALL) -m 0755 $(TARGET) $(DESTDIR)$(BINDIR)
	cp $(TARGET) ${GOPATH}/bin
