##@ The commands are:

DIR := certs

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<command>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: generate
generate: ## Generated Files
	@rm -rf ./flowedge_server
	GOOS=linux GOARCH=amd64 go build -v -o flowedge_server

.PHONY: cert
cert: ## Generate Certs
	@echo "👉 Create a certificate dir..."
	@mkdir $(DIR)
	@echo "✅ Dir complete."
	@echo "👉 Generate ca file..."
	@openssl genrsa -out ./$(DIR)/ca.key 4096
	@openssl req -x509 -new -nodes -key ./$(DIR)/ca.key -sha256 -days 3650 -out ./$(DIR)/ca.crt -subj "/CN=My-Root-CA"
	@echo "✅ Ca files complete."
	@echo "👉 Generate openssl-san.cnf file..."
	@echo "\
[req] \n\
default_bits       = 2048 \n\
prompt             = no \n\
default_md         = sha256 \n\
req_extensions     = req_ext \n\
distinguished_name = dn \n\
\n\
[dn] \n\
CN = localhost \n\
\n\
[req_ext] \n\
subjectAltName = @alt_names \n\
\n\
[alt_names] \n\
DNS.1 = localhost \n\
IP.1 = 127.0.0.1 \n\
IP.2 = 10.11.11.56 \n\
IP.3 = 47.103.98.61 \n\
" > ./$(DIR)/openssl-san.cnf
	@echo "✅ openssl-san.cnf file complete."
	@echo "👉 Generate server file..."
	@openssl genrsa -out ./$(DIR)/server.key 2048
	@openssl req -new -key ./$(DIR)/server.key -out ./$(DIR)/server.csr -config ./$(DIR)/openssl-san.cnf
	@openssl x509 -req -in ./$(DIR)/server.csr -CA ./$(DIR)/ca.crt -CAkey ./$(DIR)/ca.key -CAcreateserial -out ./$(DIR)/server.crt -days 365 -extensions req_ext -extfile ./$(DIR)/openssl-san.cnf
	@echo "✅ server file complete."
	@echo "👉 Generate client file..."
	@openssl genrsa -out ./$(DIR)/client.key 2048
	@openssl req -new -key ./$(DIR)/client.key -out ./$(DIR)/client.csr -subj "/CN=client"
	@openssl x509 -req -in ./$(DIR)/client.csr -CA ./$(DIR)/ca.crt -CAkey ./$(DIR)/ca.key -CAcreateserial -out ./$(DIR)/client.crt -days 365
	@echo "✅ client file complete."
	@tree $(DIR)
