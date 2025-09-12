# == Makefile ==

BIN_DIR := bin
TMP_DIR := tmp
GOEXE := $(shell go env GOEXE)

MERGE_FILES ?= Makefile *.go *.sh *.md

# for merging code and maybe for something else...
SRC ?= .
DST ?= 1

# Кастомные флаги сборки (можно переопределить при вызове make)
GO_BUILD_FLAGS ?=

# Находим все поддиректории в cmd, которые потенциально могут быть бинарниками
CMDS := $(wildcard cmd/*)

# Генерируем список целей для бинарников
BINARIES := $(patsubst cmd/%,$(BIN_DIR)/%,$(CMDS))


# Основная цель - собирает все бинарники
all: deps $(BINARIES)

# Правило для подготовки зависимостей
deps:
	go mod tidy
	
# Шаблонное правило для сборки любого бинарника
$(BIN_DIR)/%: FORCE
	@mkdir -p $(@D)
	go build $(GO_BUILD_FLAGS) -o $@$(GOEXE) ./cmd/$(notdir $@)

# Очистка
clean:
	-rm -rf $(BIN_DIR) $(TMP_DIR)


.PHONY: all deps clean merge-code FORCE

FORCE:


MERGE_FIND_PARTS := $(patsubst %,-o -name '%',$(MERGE_FILES))
MERGE_FIND_EXPR := $(wordlist 2,$(words $(MERGE_FIND_PARTS)),$(MERGE_FIND_PARTS))

merge-code:
	@mkdir -p $(TMP_DIR)
	find $(SRC) -type f \( $(MERGE_FIND_EXPR) \) -exec cat {} + > $(TMP_DIR)/$(DST).code
	