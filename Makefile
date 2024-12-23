
build:
	go build -o book ./cmd

fmt:
	find . -name '*.go' -exec gofumpt -w -s -extra {} \;

readme:
	code2prompt --template ~/code2prompt/templates/write-github-readme.hbs --output readme.md ./ && sed -i "s|$(HOME)||g" readme.md

prompt:
	code2prompt --template ~/code2prompt/templates/clean-up-code.hbs --output chapter.md chapter.go &&  sed -i "s|$(HOME)||g" chapter.md
	code2prompt --template ~/code2prompt/templates/clean-up-code.hbs --output compile.md compile.go &&  sed -i "s|$(HOME)||g" compile.md
	code2prompt --template ~/code2prompt/templates/clean-up-code.hbs --output render.md render.go &&  sed -i "s|$(HOME)||g" render.md
	code2prompt --template ~/code2prompt/templates/clean-up-code.hbs --output table.md table.go &&  sed -i "s|$(HOME)||g" table.md
	code2prompt --template ~/code2prompt/templates/clean-up-code.hbs --output types.md types.go &&  sed -i "s|$(HOME)||g" types.md
	code2prompt --template ~/code2prompt/templates/clean-up-code.hbs --output util.md util.go &&  sed -i "s|$(HOME)||g" util.md
