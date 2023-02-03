lint:
	( \
		which golint >/dev/null 2>&1 \
		|| ( \
			echo "Missing golint" \
			&& exit 1 \
		) \
	) \
	&& ( \
		golint ./... \
		| grep -v ^vendor \
		&& exit 1 \
		|| echo "Lint-free!" \
	)

.PHONY: lint
