##
# auth-pocketbase-attempt
#
# @file
# @version 0.1
TAILWIND_SRC = pages/input.css
TEMPLATES = $(wildcard pages/templates/*.gohtml pages/templates/**/*.gohtml)
TAILWIND_CONFIG = tailwind.config.js
TAILWIND_OUT = pages/static/public/out.css
BINARY_NAME = auth-pocketbase-attempt

.PHONY: build
build: tailwindcss
	go build -o=./${BINARY_NAME} .

.PHONY: run
run: tailwindcss
	go run . serve

# this will restart the server on source change
# and will sometimes also recompile tailwind out.css which is needed for bundling
.PHONY: run/live
run/live:
	wgo -verbose -file=.go -file=.gohtml -file=tailwind.config.js make run

# this is a phony job
# it gets executed every time it's called directly or as a dependency
# but, if out.css is fresh enough no compilation is called
.PHONY: tailwindcss
tailwindcss: $(TAILWIND_OUT)

# this is a job for producing out.css
# it's dependencies are files that should trigger compilation
# if resulting file is fresher than all of these - no build necessary
$(TAILWIND_OUT): $(TAILWIND_SRC) $(TEMPLATES) $(TAILWIND_CONFIG)
	tailwindcss -i $(TAILWIND_SRC) -o $(TAILWIND_OUT)

# end
