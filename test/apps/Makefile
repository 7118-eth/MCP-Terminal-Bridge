.PHONY: all clean echo menu progress

all: echo menu progress

echo:
	go build -o echo echo.go

menu:
	go build -o menu menu.go

progress:
	go build -o progress progress.go

clean:
	rm -f echo menu progress

run-echo: echo
	./echo

run-menu: menu
	./menu

run-progress: progress
	./progress