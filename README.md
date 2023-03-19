# Requirements Checker

![Tests](https://github.com/nrbeach/requirements_checker/actions/workflows/test.yml/badge.svg?branch=main)

---

A tool to automatically check if the modules installed in your virtual environment
match your `requirements*.txt` file(s). This is primarily intended to be used as
a pre-commit hook, but can be invoked as a CLI tool.

## Usage

### Pre-Commit

---

Add this to your `.pre-commit-config.yaml`:

```yaml
-   repo: https://github.com/nrbeach/requirements_checker
    rev: v0.1.0 # Or the particular ref
    hooks:
    -   id: requirements_checker
```

### CLI Tool

Build or install to your `$GOPATH` using `go install`. By default
`requirements_checker` will compare your virtual environment to
`requirements.txt`. Different files can be specified by the `--files` flag.


```text
# requirements_checker --files requirements.txt,requirements-dev.txt
+----------+-------------+---------+----------------------+
| MODULE   | ENVIRONMENT | DEFINED | FOUND                |
+----------+-------------+---------+----------------------+
| pathspec | 0.11.1      | 0.12.1  | requirements.txt     |
| black    | 23.1.0      | 23.1.1  | requirements-dev.txt |
| colorama | 0.4.6       | Missing | Environment          |
+----------+-------------+---------+----------------------+
```

---

## Development

This was written using Go 1.20.2. Run the `setup` make target to install the required pre-commit hooks.

---

### Contributing

Contributions are welcome. Please ensure you add relevant test cases for any changes.

1. [Fork the project](https://github.com/nrbeach/requirements_checker/fork).
1. Create your change branch (`git switch -c my-branch`)
1. Add your modifications.
1. Push to the branch (`git push origin my-branch`).
1. Create new pull request.


## FAQ

---

TODO
