.PHONY: man clean deb gem default all go

ifeq ($V, 1)
  VERBOSE =
else
  VERBOSE = @
  VQUIET = >/dev/null
  VSILENCE = >/dev/null 2>&1
endif

VERSION ?= $(shell cat VERSION)

INFO := @printf "\x1b[36m%10s\x1b[1;35m → \x1b[0;36m%s\x1b[0m\n"

## Go #########################################################################
PACKAGE    := github.com/Shopify/ecfg
GO_FILES   := $(shell find . -type f -name '*.go')

## Manpages ###################################################################
RONN ?= bundle exec ronn

RONN_FLAGS :=                         \
	--organization Shopify              \
	--manual       "Version $(VERSION)" \
	--style        toc                  \
	--warnings                          \
	--roff                              \
	--html

RONN_FILES := $(shell find man -name '*.ronn')  # man/ecfg.1.ronn
BUILD_ROFF_FILES := $(RONN_FILES:%.ronn=%)      # man/ecfg.1
BUILD_HTML_FILES := $(RONN_FILES:%.ronn=%.html) # man/ecfg.1.html

# man/ecfg.1 -> man/man1/ecfg.1
ROFF_FILES := $(foreach v, $(BUILD_ROFF_FILES), \
  $(patsubst man.%,man/man%, \
    $(subst man/,man$(suffix $v)/,$v) \
  ) \
)

# man/ecfg.1.html -> man/html/ecfg.1.html
HTML_FILES := $(foreach v, $(BUILD_HTML_FILES), \
	$(subst man/,man/html/,$v) \
)

## Debian/Ubuntu Package ######################################################
FPM ?= bundle exec fpm
DEB := target/ecfg_$(VERSION)_amd64.deb

DEB_MANIFEST := \
	usr/local/bin/ecfg \
	DEBIAN/control \
  $(foreach v, $(ROFF_FILES), $(addprefix usr/local/share/,$v))

DEB_FILES := $(foreach v, $(DEB_MANIFEST), \
	$(addprefix build/ecfg_$(VERSION)_amd64/,$v) \
)

## Rubygem ####################################################################
GEM := target/ecfg-$(VERSION).gem

GEM_MANIFEST := \
	bin/ecfg \
	build/darwin-amd64/ecfg \
	build/linux-amd64/ecfg \
	$(ROFF_FILES) \
	ecfg.gemspec \
	lib/ecfg/version.rb \
	LICENSE

GEM_FILES := \
	build/rubygem/MANIFEST \
	$(foreach v, $(GEM_MANIFEST), $(addprefix build/rubygem/,$v))

## Table of Contents ##########################################################
default: all
all:     deb man gem
man:     $(BUILD_ROFF_FILES) $(BUILD_HTML_FILES) $(ROFF_FILES) $(HTML_FILES)
deb:     $(DEB)
gem:     $(GEM)
go:      build/bin/linux-amd64 build/bin/darwin-amd64

## Manpages ###################################################################
man/man1/%: man/%
	@mkdir -p "$(@D)"
	@cp "$<" "$@"
man/man5/%: man/%
	@mkdir -p "$(@D)"
	@cp "$<" "$@"
man/html/%: man/%
	@mkdir -p "$(@D)"
	@cp "$<" "$@"

man/%: man/%.ronn
	$(INFO) 'ronn' 'man/*'
	$(VERBOSE) $(RONN) $(RONN_FLAGS) man/*.ronn $(VSILENCE)

man/%.html: man/%.ronn
	$(INFO) 'ronn' 'man/*'
	$(VERBOSE) $(RONN) $(RONN_FLAGS) man/*.ronn $(VQUIET)

## Debian/Ubuntu Package ######################################################
build/ecfg_$(VERSION)_amd64/usr/local/bin/ecfg: build/bin/linux-amd64
	@mkdir -p $(@D)
	@cp "$<" "$@"
build/ecfg_$(VERSION)_amd64/usr/local/share/man/man1/%: man/man1/%
	@mkdir -p $(@D)
	@cp "$<" "$@"
build/ecfg_$(VERSION)_amd64/usr/local/share/man/man5/%: man/man5/%
	@mkdir -p $(@D)
	@cp "$<" "$@"
build/ecfg_$(VERSION)_amd64/DEBIAN/control: dist/debian/control.tpl
	@mkdir -p $(@D)
	$(VERBOSE) sed 's/{{VERSION}}/$(VERSION)/g' < "$<" > "$@"

$(DEB): $(DEB_FILES)
	$(INFO) 'dpkg-deb' '$@'
	@mkdir -p "$(@D)"
	$(VERBOSE) $(VQUIET) dpkg-deb --build "build/ecfg_$(VERSION)_amd64" "$@"

## Rubygem ####################################################################
$(GEM): $(GEM_FILES)
	@mkdir -p $(@D)
	$(INFO) 'gem' '$(GEM)'
	$(VERBOSE) cd build/rubygem && $(VQUIET) gem build ecfg.gemspec
	@mv build/rubygem/ecfg-$(VERSION).gem $(GEM)

build/rubygem/ecfg.gemspec: dist/rubygem/ecfg.gemspec
	@mkdir -p $(@D)
	@cp "$<" "$@"
build/rubygem/bin/ecfg: dist/rubygem/bin/ecfg
	@mkdir -p $(@D)
	@cp "$<" "$@"

build/rubygem/MANIFEST: Makefile
	@mkdir -p $(@D)
	@echo $(GEM_MANIFEST) | tr ' ' \\n > $@

build/rubygem/LICENSE: LICENSE
	@mkdir -p $(@D)
	@cp "$<" "$@"

build/rubygem/man/man1/%: man/man1/%
	@mkdir -p $(@D)
	@cp "$<" "$@"
build/rubygem/man/man5/%: man/man5/%
	@mkdir -p $(@D)
	@cp "$<" "$@"

build/rubygem/build/%-amd64/ecfg: build/bin/%-amd64
	@mkdir -p $(@D)
	@cp "$<" "$@"

build/rubygem/lib/ecfg/version.rb: VERSION
	@mkdir -p $(@D)
	@echo "module Ecfg\n  VERSION = \"$(VERSION)\"\nend" > "$@"

## Go #########################################################################
build/bin/%-amd64: $(GO_FILES)
	$(INFO) 'go build' '$@'
	$(VERBOSE) $(VQUIET) \
	  env GOOS=$(subst -amd64,,$(@F)) GOARCH=amd64 \
		go build -o "$@" \
		-ldflags "-X main.version=$(VERSION)" \
		"$(PACKAGE)/cmd/ecfg"

## Misc #######################################################################
publish-site: $(HTML_FILES)
	$(INFO) 'html' '☁️'
	$(VERBOSE) dist/site/publish

clean:
	@rm -f $(BUILD_ROFF_FILES) $(BUILD_HTML_FILES)
	@rm -rf target build
