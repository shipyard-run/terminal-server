build:
	docker build -t registry.shipyard.run/terminal-server .

push:
	docker push registry.shipyard.run/terminal-server

all: build push