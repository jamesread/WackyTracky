OPENAPI := ../docs/source/api/openapi.yaml

OPENAPI_EMBED := ../service/internal/apidocs/openapi.yaml

default: generate openapi-info openapi-embed

generate:
	buf generate

# The remote connect-openapi plugin cannot set document `info`, so inject a
# valid OpenAPI `info` block (title/version are required by the 3.1 spec).
openapi-info:
	sed -i 's|^info: {}|info:\n  title: WackyTracky Client API\n  version: v1\n  description: Connect RPC API for WackyTracky. Every procedure is a POST with a JSON request/response body at /api/wackytracky.clientapi.v1.WackyTrackyClientService/<Method>. See the docs site for the human-readable reference and usage examples.|' $(OPENAPI)

openapi-embed: $(OPENAPI)
	cp $(OPENAPI) $(OPENAPI_EMBED)

.PHONY: default generate openapi-info openapi-embed
