NAME=kube-ipam
VERSION=v0.2.0

all:kube-ipam

kube-ipam:
	@echo Start building kube-ipam.
	go build -o $(NAME)
	@echo Finished building.




