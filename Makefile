rectail : cmd/rectail/args.go cmd/rectail/rectail.go
	go test -v && go install ./cmd/rectail && echo 'installed to $$GOPATH/bin/rectail' && \
	echo "to use, run 'rectail --help'"

test :
	go test -v