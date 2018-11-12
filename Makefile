export GO111MODULE=on

install:
	@echo "Installing by vendoring packages from go.mod"
	@go mod vendor
	@echo "Installation successful!"
