# githooks
There is a sample `pre-commit` which runs `go fmt` and `go test` before commit.

Copy `pre-commit` into `.git/hooks/` and make it executable (`chmod +x .git/hooks/pre-commit`) to enable this pre-commit hook.
