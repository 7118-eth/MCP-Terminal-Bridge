.PHONY: all clean echo menu progress vim

all: echo menu progress vim

echo:
	go build -o echo echo.go

menu:
	go build -o menu menu.go

progress:
	go build -o progress progress.go

vim:
	go build -o vim vim.go

clean:
	rm -f echo menu progress vim

run-echo: echo
	./echo

run-menu: menu
	./menu

run-progress: progress
	./progress

run-vim: vim
	./vim