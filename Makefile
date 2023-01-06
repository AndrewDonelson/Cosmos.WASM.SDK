DBG_MAKEFILE=1
SHELL := /bin/bash
# BuildVersion is uses the target release date from year to day seperated by DOTs (.)
# 2022.12.31
#
# following this is the character `f` which defines the number of Features & fixes seperated by DOT (.)
# f3.14
#
# this states that this release has 3 features and/or breaking changes plus 14 fixes/changes
#
# full example is
# 2022.12.31f3.14
#
# this states that the this version was released on 31st December 2022 and it contains 3 features and 14 fixes/changes
# remember to update the BUILDVERION value below this line before official releases and to set the Commit Hash tag to the same value
BUILDVERSION=2023.01.06f1.0
BUILDDATE=$(date '+%Y.%m.%d')
CHAINNAME=cosm2wasmsdk

### DO NOT EDIT BELOW THIS LINE ###
export VERSION=$(BUILDVERSION)
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_DIR := $(dir $(MKFILE_PATH))

REPOSITORY="$(basename `git rev-parse --show-toplevel`)"
@echo "Auto-Detected repository: $(REPOSITORY)"

build:
	$(info ***** building $(PROJECT_NAME) *****)
	go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILDDATE)" -o ./bin/$(PROJECT_NAME)

clean:
	$(info ***** cleaning $(PROJECT_NAME) *****)
	@rm -f ./bin/$(PROJECT_NAME)
