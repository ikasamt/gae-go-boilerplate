PROJECT_ID := XXXXXXXXX
DB_INSTANCE := XXXXXXXXX

infomation:
	$(info PROJECT_ID: $(PROJECT_ID))

gss:
	@echo "COMPILE STYLESHEETS ..."
	@docker run -i -v `PWD`/assets:/opt/assets \
		chriscannon/google-closure-tools \
		bash -c 'java -jar closure-stylesheets.jar /opt/assets/gss/*.css' \
		> public/css/all.css
	@echo "COMPILE STYLESHEETS DONE"

watch:
	@echo "WATCHING FOR CHANGES ..."
	@fswatch assets | xargs -n1 -I{} make gss

dev: infomation
	dev_appserver.py app.yaml

deploy: infomation
	gcloud app deploy --project $(PROJECT_ID)

browse: infomation
	gcloud app browse --project $(PROJECT_ID)

sqlconnect:
	gcloud sql connect $(DB_INSTANCE) --project $(PROJECT_ID)