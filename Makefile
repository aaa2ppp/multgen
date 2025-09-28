# == Makefile ==

BIN_DIR := bin
TMP_DIR := ./tmp
GOEXE := $(shell go env GOEXE)
TEST_FLAGS ?=

MERGE_FILES ?= Makefile go.mod go.sum *.go *.sh *.md

# source and destination for merge/patch operations
SRC ?= .
DST ?= 1

# Кастомные флаги сборки (можно переопределить при вызове make)
GO_BUILD_FLAGS ?=

# Находим все поддиректории в cmd, которые потенциально могут быть бинарниками
CMDS := $(wildcard cmd/*)

# Генерируем список целей для бинарников
BINARIES := $(patsubst cmd/%,$(BIN_DIR)/%,$(CMDS))


.PHONY: all deps clean merge-code patch FORCE

FORCE:


# Основная цель - собирает все бинарники
all: deps build

# Правило для подготовки зависимостей
deps:
	go mod tidy
	
# Шаблонное правило для сборки любого бинарника
$(BIN_DIR)/%: FORCE
	@mkdir -p $(@D)
	go build $(GO_BUILD_FLAGS) -o $@$(GOEXE) ./cmd/$(notdir $@)

build: $(BINARIES)

# Очистка
clean:
	-rm -rf $(BIN_DIR) $(TMP_DIR)


.PHONY: bench-http test

bench-http:
	go test -bench . -benchmem ./internal/api/... ./internal/cmd/multgen/. 

test:
	go test $(TEST_FLAGS) ./...
	@echo OK


MERGE_FIND_PARTS := $(patsubst %,-o -name '%',$(MERGE_FILES))
MERGE_FIND_EXPR := $(wordlist 2,$(words $(MERGE_FIND_PARTS)),$(MERGE_FIND_PARTS))

merge:
	@mkdir -p $(TMP_DIR)
	find $(SRC) -type f \( $(MERGE_FIND_EXPR) \) -exec cat {} + > $(TMP_DIR)/$(DST).code
	@echo "Merge saved to $(TMP_DIR)/$(DST).code"	
	

# Создает прекоммит патч
patch: test
	@mkdir -p $(TMP_DIR)
	
	@(set -e; \
	staged_list="$(TMP_DIR)/staged_list.$$$$"; \
	unstaged_list="$(TMP_DIR)/unstaged_list.$$$$"; \
	git diff --staged --name-only -- $(SRC) > "$$staged_list"; \
	git diff --name-only -- $(SRC) > "$$unstaged_list"; \
	intersection=$$(grep -Fxf "$$staged_list" "$$unstaged_list" || true); \
	rm -f "$$staged_list" "$$unstaged_list"; \
	if [ -n "$$intersection" ]; then \
		echo "" >&2; \
		echo "WARNING: the following files have changes not staged for commit:" >&2; \
		echo "  (use \"git add <file>...\" to update what will be committed)" >&2; \
		printf '%s\n' $$intersection | sed 's/^/        /' >&2; \
		echo "" >&2; \
	fi)
	
	git diff --staged -- $(SRC) > $(TMP_DIR)/$(DST).patch
	@echo "Patch saved to $(TMP_DIR)/$(DST).patch"
