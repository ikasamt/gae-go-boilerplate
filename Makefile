PROJECT_ID := XXXXXXXXX

infomation:
	$(info PROJECT_ID: $(PROJECT_ID))

gss:
	@echo "COMPILE STYLESHEETS START..."
	@docker run -i -v `PWD`/assets:/opt/assets \
		chriscannon/google-closure-tools \
		bash -c 'java -jar closure-stylesheets.jar /opt/assets/gss/*.css' \
		> public/css/all.css
	@echo "COMPILE STYLESHEETS DONE"

build: gss genny
	@echo ""

watch:
	@echo "WATCHING FOR CHANGES ..."
	@fswatch assets clefs | xargs -n1 -I{} make build

dev: infomation
	dev_appserver.py app.yaml 2>&1 | go-colorize

deploy: infomation
	@echo "gcloud app deploy --project $(PROJECT_ID)"

browse: infomation
	@echo "gcloud app browse --project $(PROJECT_ID)"

sqlconnect:
	@echo "MEMO: gcloud sql connect $(DB_INSTANCE) --project $(PROJECT_ID) "

clean:
	rm -rf zzz-autogen-*

genny: clean
	@echo "GENNY START..."
	genny -pkg=app -in=./clefs/find.go -out=`PWD`/zzz-find.go gen "Anything=User,Organization,PasswordReset"
	@echo "GENNY DONE"
