all:
	go build -o bin/mvela main.go

uninstall:
	docker ps |grep k3d-mvela-cluster |grep '^............'  -o|xargs -L 1 docker kill