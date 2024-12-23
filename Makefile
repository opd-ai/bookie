
build:
	go build -o book ./cmd

fmt:
	find . -name '*.go' -exec gofumpt -w -s -extra {} \;

readme:
	code2prompt --template ~/code2prompt/templates/write-github-readme.hbs --output example/readme.md ./ && sed -i "s|$(HOME)||g" example/readme.md

prompt:
	code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output example/chapter.md chapter.go &&  sed -i "s|$(HOME)||g" example/chapter.md
	code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output example/compile.md compile.go &&  sed -i "s|$(HOME)||g" example/compile.md
	code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output example/render.md render.go &&  sed -i "s|$(HOME)||g" example/render.md
	code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output example/rendercore.md rendercore.go &&  sed -i "s|$(HOME)||g" example/rendercore.md
	code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output example/rendersub.md rendersub.go &&  sed -i "s|$(HOME)||g" example/rendersub.md
	code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output example/table.md table.go &&  sed -i "s|$(HOME)||g" example/table.md
	code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output example/types.md types.go &&  sed -i "s|$(HOME)||g" example/types.md
	code2prompt --template ~/code2prompt/templates/document-the-code.hbs --output example/util.md util.go &&  sed -i "s|$(HOME)||g" example/util.md
