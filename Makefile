all: build run

build:
	go build -o ./dist/ssh-lxd

run:
	./dist/ssh-lxd
