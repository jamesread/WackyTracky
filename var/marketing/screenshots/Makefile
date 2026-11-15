# Regenerates marketing screenshots against a throwaway WackyTracky server
# seeded with fictional todo.txt data. Run `make` here; everything is torn
# down afterwards.
#
# The todo.txt backend needs no database container, so the "stack" is just
# the service binary, run from tmp/run with its own config, data and HOME (so
# git sync SSH key setup never touches ~/.ssh or the global git config).

ROOT := $(abspath ../../..)
HERE := $(abspath .)

DOCS_IMAGES := $(ROOT)/docs/modules/ROOT/images/marketing
PNGS := home.png list.png task-details.png search.png tag-properties.png

PORT := 18480
BACKEND_URL := http://127.0.0.1:$(PORT)

.PHONY: all build up seed shoot manifest publish down clean

all:
	@$(MAKE) build seed up shoot manifest publish; rc=$$?; $(MAKE) down; exit $$rc

build:
	[ -d $(ROOT)/frontend/node_modules ] || (cd $(ROOT)/frontend && npm ci)
	$(MAKE) -C $(ROOT)/frontend build
	$(MAKE) -C $(ROOT)/service build

# The todo.txt backend reads its files at startup, so seeding is a fresh copy of
# the committed fixtures made before the server starts. Copying (rather than
# creating through the API) keeps the fixed list/task IDs, and so stable URLs.
seed: down
	rm -rf tmp/run
	mkdir -p tmp/run/service/data tmp/run/home
	./render-seed.py tmp/run/service/data/todotxt
	cp service.yaml tmp/run/service/config.yaml
	# The service finds its UI at ../frontend/dist relative to its working dir.
	ln -sfn $(ROOT)/frontend tmp/run/frontend

up:
	cd tmp/run/service && { HOME=$(HERE)/tmp/run/home PORT=$(PORT) WACKYTRACKY_SECURE_COOKIES=false \
		WACKYTRACKY_MIGRATIONS_DIR=$(ROOT)/database/sqlite/migrations \
		$(ROOT)/service/wacky-tracky-server > $(HERE)/tmp/service.log 2>&1 & echo $$! > $(HERE)/tmp/service.pid; }
	@for i in $$(seq 60); do curl -fsS -o /dev/null $(BACKEND_URL)/llms.txt && exit 0; sleep 0.5; done; \
		echo "Backend did not start; see tmp/service.log"; exit 1

shoot:
	repo-helper screenshot --config screenshots.ini

manifest:
	./generate-manifest.py

# The README references the PNGs here directly; Antora only serves images from
# the docs module, so they are copied there.
publish:
	mkdir -p $(DOCS_IMAGES)
	cp $(PNGS) $(DOCS_IMAGES)/

down:
	-@[ -f tmp/service.pid ] && kill $$(cat tmp/service.pid) 2>/dev/null; rm -f tmp/service.pid

clean: down
	rm -rf tmp
