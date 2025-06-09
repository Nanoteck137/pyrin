clean:
	rm -rf work
	rm -rf tmp
	rm -f result

test-build:
	nix build --no-link .#

publish: test-build
	publish-version

.PHONY: clean publish
