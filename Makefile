export GO111MODULE=on

install:
	@echo "Installing by vendoring packages from go.mod"
	@sudo apt-get install poppler-utils wv unrtf tidy
	@go mod vendor
	@echo "Installation successful!"
