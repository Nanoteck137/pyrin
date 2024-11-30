clean:
	rm -rf work
	rm -rf tmp
	rm -f result

publish:
	nix build --no-link .#
	publish-version

.PHONY: clean publish
