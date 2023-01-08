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
NUMMAJOR=0
NUMMINOR=1
PROJECTNAME=cosm2wasmsdk

### DO NOT EDIT BELOW THIS LINE ###
BUILDDATE=$(date '+%Y.%m.%d')
BUILDVERSION=$(BUILDDATE)f$(NUMMAJOR).$(NUMMINOR)
export VERSION=$(BUILDVERSION)
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_DIR := $(dir $(MKFILE_PATH))

REPOSITORY="$(basename `git rev-parse --show-toplevel`)"
@echo "Auto-Detected repository: $(REPOSITORY)"

build:
	$(info ***** building $(PROJECTNAME) *****)
	@cd $(PROJECT_DIR)/cmd && go build -ldflags "-s -w -X main.Version=$(BUILDVERSION) -X main.BuildDate=$(BUILDDATE)" -o $(PROJECT_DIR)/bin/$(PROJECTNAME) && cd $(PROJECT_DIR)
	$(info ***** done *****)

clean:
	$(info ***** cleaning $(PROJECTNAME) *****)
	@rm -f $(PROJECT_DIR)/bin/$(PROJECTNAME)


build-wasm:
	$(info ***** building $(PROJECTNAME) wasm sdk *****)
	@cd $(PROJECT_DIR)/sdk && GOOS=js GOARCH=wasm go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildDate=$(LASTBUILDDATE)" -o $(PROJECT_DIR)/bin/$(PROJECTNAME).wasm && cd $(PROJECT_DIR)

clean-wasm:
	$(info ***** cleaning $(PROJECT_NAME) wasm sdk *****)
	@rm -f ./bin/$(PROJECT_NAME).wasm
