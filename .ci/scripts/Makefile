##############################################################################
# Variables used for various build targets.
##############################################################################

BUILD=../../build
PACKAGE_STORAGE=.package-storage
PACKAGE_STORAGE_PATH=$(BUILD)/$(PACKAGE_STORAGE)
PACKAGE_STORAGE_BRANCH=update-apm-$(shell date "+%Y%m%d%H%M%S")
APM_SERVER_VERSION=$(shell make --no-print-directory -C ../../ get-version)

##############################################################################
# Rules for the package-storage.
##############################################################################
## package-storage-snapshot : Clone the package-storage@snapshot and copy the apmpachage folder
.PHONY: package-storage-snapshot
package-storage-snapshot:
	@rm -fr $(PACKAGE_STORAGE_PATH)
	git clone https://github.com/elastic/package-storage.git $(PACKAGE_STORAGE_PATH) --branch snapshot --single-branch --depth=1
	cp -rf $(BUILD)/integrations/apm/$(APM_SERVER_VERSION) $(PACKAGE_STORAGE_PATH)/packages/apm/

## create-package-storage-pull-request : Create the pull request for the package storage
.PHONY: create-package-storage-pull-request
create-package-storage-pull-request:
	@cd $(PACKAGE_STORAGE_PATH) ; \
		echo "INFO: create branch" ; \
		git checkout -b $(PACKAGE_STORAGE_BRANCH) ; \
		echo "INFO: add files if any" ; \
		git add . ; \
		echo "INFO: any changes" ; \
		rm .changes 2>/dev/null || true ; \
		git diff --staged --quiet || touch .changes ; \
		echo "INFO: commit changes if any" ; \
		test -f .changes && git commit -m "[automation] Publish apm-$(APM_SERVER_VERSION)" || true ; \
		echo "INFO: show remote details" ; \
		git remote -v ; \
		echo "INFO: push branch if changes" ; \
		test -f .changes && git push --set-upstream origin $(PACKAGE_STORAGE_BRANCH) || true ; \
		echo "INFO: create pull request if changes" ; \
		test -f .changes && gh pr create \
			--title "Publish apm-$(APM_SERVER_VERSION)" \
			--body "Automated by $${BUILD_URL}" \
			--label automation \
			--base snapshot \
			--head $(PACKAGE_STORAGE_BRANCH) \
			--reviewer elastic/apm-server || true ; \

## get-package-storage-location : Get the package storage location
.PHONY: get-package-storage-location
get-package-storage-location:
	@echo "build/$(PACKAGE_STORAGE)"
