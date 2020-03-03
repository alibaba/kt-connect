go_build:
	go build

test: go_build prune_test
	
prune_test:
	sudo ./ktctl connect --method=vpn