build:
	docker build -t shipyardrun/terminal-server .

push:
	docker push shipyardrun/terminal-server

all: build push